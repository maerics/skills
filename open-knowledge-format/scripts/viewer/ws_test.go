package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestLifecycleShutsDownAfterGracePeriod(t *testing.T) {
	done := make(chan struct{})
	lc := newLifecycle(30*time.Millisecond, time.Hour, func() { close(done) })
	lc.onConnect()
	lc.onDisconnect()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shutdown was not called after the grace period elapsed with zero clients")
	}
}

func TestLifecycleGraceCancelledByReconnect(t *testing.T) {
	var called int32
	lc := newLifecycle(80*time.Millisecond, time.Hour, func() { atomic.AddInt32(&called, 1) })
	lc.onConnect()
	lc.onDisconnect() // starts the grace timer
	time.Sleep(30 * time.Millisecond)
	lc.onConnect() // reconnect within the grace window cancels it
	time.Sleep(120 * time.Millisecond)
	if atomic.LoadInt32(&called) != 0 {
		t.Fatal("shutdown fired despite a reconnect within the grace period")
	}
}

func TestLifecycleIdleTimeoutBeforeAnyConnection(t *testing.T) {
	done := make(chan struct{})
	newLifecycle(time.Hour, 30*time.Millisecond, func() { close(done) })
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("shutdown was not called after the idle timeout with no connections ever")
	}
}

func TestLifecycleIdleTimeoutCancelledByConnect(t *testing.T) {
	var called int32
	lc := newLifecycle(time.Hour, 40*time.Millisecond, func() { atomic.AddInt32(&called, 1) })
	time.Sleep(10 * time.Millisecond)
	lc.onConnect()
	time.Sleep(80 * time.Millisecond)
	if atomic.LoadInt32(&called) != 0 {
		t.Fatal("idle timeout fired despite a connection arriving before it elapsed")
	}
}

// TestServeWSHandshakeAndDisconnect drives serveWS over a real TCP
// connection: performs the RFC 6455 upgrade handshake by hand, confirms the
// lifecycle sees the connection, then closes the socket and confirms both
// the disconnect and the resulting grace-period shutdown fire.
func TestServeWSHandshakeAndDisconnect(t *testing.T) {
	shutdown := make(chan struct{})
	lc := newLifecycle(30*time.Millisecond, time.Hour, func() { close(shutdown) })

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) { serveWS(w, r, lc) })
	ts := httptest.NewServer(mux)
	defer ts.Close()

	addr := strings.TrimPrefix(ts.URL, "http://")
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	req := "GET /ws HTTP/1.1\r\n" +
		"Host: " + addr + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n" +
		"Sec-WebSocket-Version: 13\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read handshake response: %v", err)
	}
	resp := string(buf[:n])
	if !strings.Contains(resp, "101") {
		t.Fatalf("expected a 101 Switching Protocols response, got: %q", resp)
	}
	if !strings.Contains(resp, "Sec-WebSocket-Accept:") {
		t.Fatalf("expected a Sec-WebSocket-Accept header, got: %q", resp)
	}

	waitUntil(t, time.Second, func() bool {
		lc.mu.Lock()
		defer lc.mu.Unlock()
		return lc.count == 1
	}, "lifecycle never saw the connection")

	conn.Close()

	select {
	case <-shutdown:
	case <-time.After(time.Second):
		t.Fatal("closing the socket never led to a shutdown after the grace period")
	}
}

func waitUntil(t *testing.T, timeout time.Duration, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal(msg)
}
