package airplay2

import (
	"testing"
)

func TestClient(t *testing.T) {
	_, err := NewClient(ClientDiscoveryParameters{
		DeviceName: "test",
	})
	if err != nil {
		t.Fatalf("error creating new client: %v\n", err)
	}
}
