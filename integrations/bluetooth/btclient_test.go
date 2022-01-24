package bluetooth

import (
	"ledfx/config"
	log "ledfx/logger"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	if _, err := log.Init(config.Config{
		Verbose: true,
	}); err != nil {
		t.Fatalf("Error initializing logger: %v\n", err)
	}

	// Initialize a Bluetooth client adapter
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Error generating new BLE client: %v\n", err)
	}

	// Search through the device cache and discovery events for
	// a bluetooth device
	if err := client.SearchAndConnect(SearchTargetConfig{
		DeviceRegex:          `(?i)k850$`, // Case-insensitive, searches for "k850" ([AV] Samsung Sound Bar K850)
		ConnectRetryCoolDown: 2 * time.Second,
	}); err != nil {
		t.Fatalf("Error searching and connecting for BLE device: %v\n", err)
	}

	// Wait until a device is successfully connected
	client.WaitConnect()
}
