package bridgeapi

import (
	"ledfx/audio"
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	mux := http.DefaultServeMux
	if _, err := NewServer(func(buf audio.Buffer) {
		// Do nothing
	}, mux); err != nil {
		t.Fatalf("Error creating new server: %s", err)
	}

	if err := http.ListenAndServe("127.0.0.1:8080", mux); err != nil {
		t.Fatalf("Error listening: %s", err)
	}
}
