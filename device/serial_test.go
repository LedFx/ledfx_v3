package device

import (
	"log"
	"testing"

	"go.bug.st/serial"
)

func TestSerial(t *testing.T) {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	for _, port := range ports {
		t.Logf("Found port: %v\n", port)
	}
	t.Fail()
}
