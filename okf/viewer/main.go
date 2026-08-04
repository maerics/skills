// Command okf-viewer serves an OKF knowledge bundle as a local, navigable
// website, launches a browser at it, and shuts itself down once nobody has
// it open anymore — nothing for the caller to clean up.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

const (
	defaultGracePeriod = 3 * time.Second
	defaultIdleTimeout = 5 * time.Minute
)

func main() {
	port := flag.Int("port", 0, "port to listen on (0 = OS-assigned)")
	noOpen := flag.Bool("no-open", false, "don't launch a browser automatically")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-port N] [-no-open] [bundle-path]\n\n"+
			"Serve an OKF knowledge bundle as a local, navigable website and open\n"+
			"it in a browser. bundle-path defaults to \".\"; if it contains a .okf\n"+
			"subdirectory, that is used as the bundle root. The server shuts\n"+
			"itself down once no browser tab has it open, so there's nothing to\n"+
			"clean up manually.\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	bundlePath := "."
	if flag.NArg() > 0 {
		bundlePath = flag.Arg(0)
	}

	root, err := resolveBundleRoot(bundlePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "okf-viewer:", err)
		os.Exit(1)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "okf-viewer:", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var once sync.Once
	shutdown := func() { once.Do(cancel) }

	lc := newLifecycle(defaultGracePeriod, defaultIdleTimeout, shutdown)
	srv := newServer(root, lc)
	httpServer := &http.Server{Handler: srv.mux()}

	go func() {
		if err := httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, "okf-viewer:", err)
		}
	}()

	url := fmt.Sprintf("http://%s/", ln.Addr().String())
	fmt.Fprintf(os.Stderr, "okf-viewer: serving %s at %s (Ctrl+C to stop, or just close the browser tab)\n", root, url)

	if !*noOpen {
		if err := openBrowser(url); err != nil {
			fmt.Fprintf(os.Stderr, "okf-viewer: couldn't open a browser automatically (%v); open %s yourself\n", err, url)
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sig:
	case <-ctx.Done():
	}

	shutdownCtx, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()
	_ = httpServer.Shutdown(shutdownCtx)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
