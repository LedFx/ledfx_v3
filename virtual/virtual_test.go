package virtual

import (
	"ledfx/color"
	"ledfx/device"
	"ledfx/effect"
	"testing"
)

func TestVirtual(t *testing.T) {

	bdc := device.BaseDeviceConfig{
		PixelCount: 64,
		Name:       "Spotlight",
	}
	udpc := device.UDPConfig{
		NetworkerConfig: device.NetworkerConfig{
			IP:   "192.168.0.72",
			Port: 21324,
		},
		Protocol: device.DRGB,
		Timeout:  60,
	}
	d, _, err := device.New("udp", bdc, udpc)
	if err != nil {
		t.Error(err)
	}
	err = d.Connect()
	if err != nil {
		t.Error(err)
	}

	e, _, err := effect.New("energy", bdc.PixelCount, nil)
	if err != nil {
		t.Error(err)
	}

	p := make(color.Pixels, bdc.PixelCount)

	e.Render(p)
	err = d.Send(p)
	if err != nil {
		t.Error(err)
	}
}
