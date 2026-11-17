package cli

import (
	"path/filepath"
	"testing"
	"time"
)

// The proxy runs inside the host app or its VPN extension, so none of these may end the process.

func TestInitLoggerFallsBackToStdoutWhenTheLogFileCantBeOpened(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-directory", "proxy.log")

	InitLogger(false, missing)

	if Logger == nil {
		t.Fatal("Logger is nil after a log file that can't be opened")
	}
	Logger.Info("still logging")
}

func TestRunRejectsAnInvalidTunnelType(t *testing.T) {
	InitLogger(false, "")
	client := NewHTTPClient("127.0.0.1:0", "wss://127.0.0.1:1", 3, 1500, func(int) {}, make(chan string, 1), false, "")

	if err := client.Run(); err == nil {
		t.Fatal("Run accepted tunnel type 3")
	}
}

func TestGoSafelyRecoversAPanic(t *testing.T) {
	InitLogger(false, "")
	done := make(chan struct{})

	goSafely("test", func() {
		defer close(done)
		panic("boom")
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("the goroutine didn't finish")
	}
}

func TestRunIgnoresAStopSentBeforeItStarted(t *testing.T) {
	InitLogger(false, "")
	channel := make(chan string, 1)
	channel <- "done"
	client := NewHTTPClient("127.0.0.1:0", "wss://127.0.0.1:1", WSTunnel, 1500, func(int) {}, channel, false, "")
	finished := make(chan error, 1)

	go func() { finished <- client.Run() }()

	select {
	case <-finished:
		t.Fatal("Run stopped on a stop sent before it started")
	case <-time.After(300 * time.Millisecond):
	}
	channel <- "done"
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("Run didn't stop")
	}
}
