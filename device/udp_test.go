package device

import (
	"ledfx/color"
	"ledfx/config"
	"ledfx/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup(pixelCount int, protocol string) (UDPDevice, []color.Color) {
	var deviceConfig = config.DeviceConfig{
		UdpPacketType: protocol,
	}
	var testDevice = UDPDevice{
		Config: deviceConfig,
	}

	data := []color.Color{}
	for i := 0; i < pixelCount; i++ {
		newColor, err := color.NewColor(color.LedFxColors["red"])
		data = append(data, newColor)
		if err != nil {
			logger.Logger.Fatal(err)
		}
	}
	return testDevice, data
}

func TestBuildWARLSPacket(t *testing.T) {
	// one LED
	testDevice, data := setup(1, "WARLS")
	packet, err := testDevice.BuildPacket(data, 0)
	if err != nil {
		t.Error(err)
	}
	expected := []byte{0x01, 0x01, 0x00, 0xff, 0x00, 0x00}
	if !assert.ElementsMatch(t, packet, expected) {
		t.Errorf("Expected %v, got %v", expected, packet)
	}

	// multiple LEDs with offset
	testDevice, data = setup(3, "WARLS")
	packet, err = testDevice.BuildPacket(data, 1)
	if err != nil {
		t.Error(err)
	}
	expected = []byte{0x01, 0x01, 0x01, 0xff, 0x00, 0x00, 0xff, 0x00, 0x00, 0xff, 0x00, 0x00}
	if !assert.ElementsMatch(t, packet, expected) {
		t.Errorf("Expected %v, got %v", expected, packet)
	}

	// too many LEDs
	testDevice, data = setup(256, "WARLS")
	packet, err = testDevice.BuildPacket(data, 0)
	if _, ok := err.(*InvalidLedCountError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Count", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}

	// negative offset
	testDevice, data = setup(1, "WARLS")
	packet, err = testDevice.BuildPacket(data, -1)
	if _, ok := err.(*InvalidLedOffsetError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Offset", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}

	// out of range offset
	testDevice, data = setup(1, "WARLS")
	packet, err = testDevice.BuildPacket(data, 256)
	if _, ok := err.(*InvalidLedOffsetError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Offset", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}
}

func TestBuildDRGBPacket(t *testing.T) {
	// one LED
	testDevice, data := setup(1, "DRGB")
	packet, err := testDevice.BuildPacket(data, 0)
	if err != nil {
		t.Error(err)
	}
	expected := []byte{0x02, 0x01, 0xff, 0x00, 0x00}
	if !assert.ElementsMatch(t, packet, expected) {
		t.Errorf("Expected %v, got %v", expected, packet)
	}

	// multiple LEDs
	testDevice, data = setup(3, "DRGB")
	packet, err = testDevice.BuildPacket(data, 0)
	if err != nil {
		t.Error(err)
	}
	expected = []byte{0x02, 0x01, 0xff, 0x00, 0x00, 0xff, 0x00, 0x00, 0xff, 0x00, 0x00}
	if !assert.ElementsMatch(t, packet, expected) {
		t.Errorf("Expected %v, got %v", expected, packet)
	}

	// too many LEDs
	testDevice, data = setup(491, "DRGB")
	packet, err = testDevice.BuildPacket(data, 0)
	if _, ok := err.(*InvalidLedCountError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Count", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}

	// negative offset
	testDevice, data = setup(1, "DRGB")
	packet, err = testDevice.BuildPacket(data, -1)
	if _, ok := err.(*InvalidLedOffsetError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Offset", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}

	// out of range offset
	testDevice, data = setup(1, "DRGB")
	packet, err = testDevice.BuildPacket(data, 1)
	if _, ok := err.(*InvalidLedOffsetError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Offset", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}
}

func TestBuildDRGBWPacket(t *testing.T) {
	// one LED
	testDevice, data := setup(1, "DRGBW")
	packet, err := testDevice.BuildPacket(data, 0)
	if err != nil {
		t.Error(err)
	}
	expected := []byte{0x03, 0x01, 0xff, 0x00, 0x00, 0x00}
	if !assert.ElementsMatch(t, packet, expected) {
		t.Errorf("Expected %v, got %v", expected, packet)
	}

	// multiple LEDs
	testDevice, data = setup(3, "DRGBW")
	packet, err = testDevice.BuildPacket(data, 0)
	if err != nil {
		t.Error(err)
	}
	expected = []byte{0x03, 0x01, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00}
	if !assert.ElementsMatch(t, packet, expected) {
		t.Errorf("Expected %v, got %v", expected, packet)
	}

	// too many LEDs
	testDevice, data = setup(491, "DRGBW")
	packet, err = testDevice.BuildPacket(data, 0)
	if _, ok := err.(*InvalidLedCountError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Count", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}

	// negative offset
	testDevice, data = setup(1, "DRGBW")
	packet, err = testDevice.BuildPacket(data, -1)
	if _, ok := err.(*InvalidLedOffsetError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Offset", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}

	// out of range offset
	testDevice, data = setup(1, "DRGBW")
	packet, err = testDevice.BuildPacket(data, 1)
	if _, ok := err.(*InvalidLedOffsetError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Offset", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}
}

func TestBuildDNRGBPacket(t *testing.T) {
	// one LED
	testDevice, data := setup(1, "DNRGB")
	packet, err := testDevice.BuildPacket(data, 0)
	if err != nil {
		t.Error(err)
	}
	expected := []byte{0x04, 0x01, 0x00, 0x00, 0xff, 0x00, 0x00}
	if !assert.ElementsMatch(t, packet, expected) {
		t.Errorf("Expected %v, got %v", expected, packet)
	}

	// multiple LEDs with offset
	testDevice, data = setup(3, "DNRGB")
	packet, err = testDevice.BuildPacket(data, 2047)
	if err != nil {
		t.Error(err)
	}
	expected = []byte{0x04, 0x01, 0x07, 0xff, 0xff, 0x00, 0x00, 0xff, 0x00, 0x00, 0xff, 0x00, 0x00}
	if !assert.ElementsMatch(t, packet, expected) {
		t.Errorf("Expected %v, got %v", expected, packet)
	}

	// too many LEDs
	testDevice, data = setup(490, "DNRGB")
	packet, err = testDevice.BuildPacket(data, 0)
	if _, ok := err.(*InvalidLedCountError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Count", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}

	// negative offset
	testDevice, data = setup(1, "DNRGB")
	packet, err = testDevice.BuildPacket(data, -1)
	if _, ok := err.(*InvalidLedOffsetError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Offset", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}

	// out of range offset
	testDevice, data = setup(1, "DNRGB")
	packet, err = testDevice.BuildPacket(data, 65536)
	if _, ok := err.(*InvalidLedOffsetError); !ok {
		t.Errorf("Expected %v, got %v", "Invalid Led Offset", err)
	} else if packet != nil {
		t.Errorf("Expected %v, got %v", nil, packet)
	}
}
