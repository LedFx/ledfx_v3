package loopback

import (
	"os"
	"testing"
	"time"
)

func TestNewLoopback(t *testing.T) {
	fi, err := os.Create("audio.pcm")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		fi.Sync()
		fi.Close()
	}()

	loop, err := New(fi)
	if err != nil {
		t.Fatalf("error during loopback init: %v\n", err)
	}

	go func() {
		time.Sleep(100 * time.Second)
		loop.Done()
	}()

	loop.Wait()
}
