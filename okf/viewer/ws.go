// A minimal RFC 6455 websocket server, just enough to know whether at
// least one browser tab currently has the viewer open. No application
// messages are expected in either direction; the browser holds one socket
// open for as long as its page is, and the server only needs connect/
// disconnect events to drive shutdown timing (see lifecycle below).
package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

const (
	opContinuation = 0x0
	opText         = 0x1
	opBinary       = 0x2
	opClose        = 0x8
	opPing         = 0x9
	opPong         = 0xA
)

// maxFrameSize bounds a single incoming frame; the viewer never expects
// anything but tiny control frames from the browser.
const maxFrameSize = 1 << 20

var errBadFrame = errors.New("malformed websocket frame")

// lifecycle tracks connected viewer sessions and calls shutdown once it's
// confident nobody is looking anymore:
//   - after the last connection closes, a short grace period rides out the
//     normal zero-clients gap during page-to-page navigation
//   - before any connection ever arrives, a longer idle timeout catches the
//     case where the browser never opened (headless run, auto-open failed,
//     or the user just never looked) so the process never lingers forever
type lifecycle struct {
	mu    sync.Mutex
	count int

	gracePeriod time.Duration
	idleTimeout time.Duration
	shutdown    func()

	idleTimer     *time.Timer
	shutdownTimer *time.Timer
}

func newLifecycle(gracePeriod, idleTimeout time.Duration, shutdown func()) *lifecycle {
	l := &lifecycle{gracePeriod: gracePeriod, idleTimeout: idleTimeout, shutdown: shutdown}
	l.idleTimer = time.AfterFunc(idleTimeout, l.fireIfStillEmpty)
	return l
}

func (l *lifecycle) fireIfStillEmpty() {
	l.mu.Lock()
	empty := l.count == 0
	l.mu.Unlock()
	if empty {
		l.shutdown()
	}
}

func (l *lifecycle) onConnect() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.count++
	if l.idleTimer != nil {
		l.idleTimer.Stop()
		l.idleTimer = nil
	}
	if l.shutdownTimer != nil {
		l.shutdownTimer.Stop()
		l.shutdownTimer = nil
	}
}

func (l *lifecycle) onDisconnect() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.count > 0 {
		l.count--
	}
	if l.count == 0 {
		l.shutdownTimer = time.AfterFunc(l.gracePeriod, l.fireIfStillEmpty)
	}
}

// serveWS upgrades r to a websocket connection and blocks until it closes,
// tracking the session against lc for the process's lifetime.
func serveWS(w http.ResponseWriter, r *http.Request, lc *lifecycle) {
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" || !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		http.Error(w, "expected websocket upgrade", http.StatusBadRequest)
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websockets unsupported", http.StatusInternalServerError)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		return
	}
	defer conn.Close()

	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + computeAccept(key) + "\r\n\r\n"
	if _, err := buf.WriteString(resp); err != nil {
		return
	}
	if err := buf.Flush(); err != nil {
		return
	}

	lc.onConnect()
	defer lc.onDisconnect()

	readLoop(conn, buf.Reader)
}

func computeAccept(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte(wsGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func readLoop(conn net.Conn, r *bufio.Reader) {
	for {
		opcode, payload, err := readFrame(r)
		if err != nil {
			return
		}
		switch opcode {
		case opClose:
			_ = writeFrame(conn, opClose, nil)
			return
		case opPing:
			if writeFrame(conn, opPong, payload) != nil {
				return
			}
		case opText, opBinary, opContinuation, opPong:
			// No application-level messages are expected; ignore.
		default:
			return
		}
	}
}

func readFrame(r *bufio.Reader) (opcode byte, payload []byte, err error) {
	var head [2]byte
	if _, err = io.ReadFull(r, head[:]); err != nil {
		return 0, nil, err
	}
	opcode = head[0] & 0x0f
	masked := head[1]&0x80 != 0
	length := int64(head[1] & 0x7f)

	switch length {
	case 126:
		var ext [2]byte
		if _, err = io.ReadFull(r, ext[:]); err != nil {
			return 0, nil, err
		}
		length = int64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err = io.ReadFull(r, ext[:]); err != nil {
			return 0, nil, err
		}
		length = int64(binary.BigEndian.Uint64(ext[:]))
	}
	if length < 0 || length > maxFrameSize {
		return 0, nil, errBadFrame
	}

	// RFC 6455 §5.1: frames from client to server MUST be masked.
	if !masked {
		return 0, nil, errBadFrame
	}
	var maskKey [4]byte
	if _, err = io.ReadFull(r, maskKey[:]); err != nil {
		return 0, nil, err
	}

	payload = make([]byte, length)
	if _, err = io.ReadFull(r, payload); err != nil {
		return 0, nil, err
	}
	for i := range payload {
		payload[i] ^= maskKey[i%4]
	}
	return opcode, payload, nil
}

func writeFrame(w io.Writer, opcode byte, payload []byte) error {
	head := []byte{0x80 | opcode} // FIN=1, no fragmentation
	n := len(payload)
	switch {
	case n <= 125:
		head = append(head, byte(n))
	case n <= 0xffff:
		var ext [2]byte
		binary.BigEndian.PutUint16(ext[:], uint16(n))
		head = append(head, 126)
		head = append(head, ext[:]...)
	default:
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(n))
		head = append(head, 127)
		head = append(head, ext[:]...)
	}
	if _, err := w.Write(head); err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	_, err := w.Write(payload)
	return err
}
