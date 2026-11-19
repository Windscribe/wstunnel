package main

import (
	"testing"
	"time"
)

func TestStopNeverBlocks(t *testing.T) {
	returned := make(chan struct{})

	go func() {
		// Nothing is running, so nothing receives: Stop must still return each time.
		Stop()
		Stop()
		close(returned)
	}()

	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("Stop blocked")
	}
}
