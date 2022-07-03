package device

import (
	"testing"

	"go.bug.st/serial"
)

func TestSerial(t *testing.T) {
	ports, err := serial.GetPortsList()
	if err != nil {
		t.Error(err)
	}
	if len(ports) == 0 {
		t.Error("No serial ports found!")
	}
	for _, port := range ports {
		t.Logf("Found port: %v\n", port)
	}
}
