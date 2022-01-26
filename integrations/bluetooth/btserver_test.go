package bluetooth

import (
	"ledfx/config"
	"testing"

	log "ledfx/logger"
)

func init() {
	log.Init(&config.Config{Verbose: true})
}

func TestNewServer(t *testing.T) {
	server, err := NewServer("LedFX")
	if err != nil {
		t.Fatalf("Error initializing Bluetooth server: %v\n", err)
	}
	if err := server.Serve(); err != nil {
		t.Fatalf("Error serving Bluetooth advertisement: %v\n", err)
	}
	defer server.CloseApp()

	server.Wait()
}
