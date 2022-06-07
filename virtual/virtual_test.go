package virtual

import (
	"fmt"
	"ledfx/audio"
	"ledfx/audio/audiobridge"
	"ledfx/color"
	"ledfx/device"
	"ledfx/effect"
	"log"
	"testing"
	"time"
)

func TestVirtual(t *testing.T) {

	bdc := device.BaseDeviceConfig{
		PixelCount: 64,
		Name:       "Spotlight",
	}
	udpc := device.UDPConfig{
		NetworkerConfig: device.NetworkerConfig{
			IP:   "192.168.0.104",
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
	ec := map[string]interface{}{
		"intensity":      1,
		"brightness":     0.5,
		"saturation":     1,
		"palette":        "Rainbow",
		"blur":           1,
		"decay":          0.5,
		"flip":           false,
		"mirror":         false,
		"bkg_brightness": 1,
		"bkg_color":      "#000000",
		"hue_shift":      0,
	}
	e, _, err := effect.New("energy", bdc.PixelCount, ec)
	if err != nil {
		t.Error(err)
	}

	p := make(color.Pixels, bdc.PixelCount)

	br, err := audiobridge.NewBridge(audio.Analyzer.BufferCallback)
	if err != nil {
		log.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	audiodevice, err := audio.GetDeviceByID("9f012a5ef29af5e7b226bae734a8cb2ad229f063") // get from config
	if err != nil {
		log.Fatal(err)
	}
	if err := br.StartLocalInput(audiodevice, true); err != nil {
		log.Fatalf("Error starting local input: %v\n", err)

	}

	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()
	done := make(chan bool)

	go func() {
		time.Sleep(29 * time.Second)
		done <- true
	}()

	for {
		select {
		case <-ticker.C:
			e.Render(p)
			err = d.Send(p)
			if err != nil {
				t.Error(err)
			}
		case <-done:
			return
		}
	}
}

func BenchmarkVirtual(t *testing.B) {

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

	ec := map[string]interface{}{
		"intensity":      1,
		"brightness":     1,
		"saturation":     1,
		"palette":        "Rainbow",
		"blur":           1,
		"flip":           true,
		"mirror":         true,
		"bkg_brightness": 1,
		"bkg_color":      "#000000",
		"hue_shift":      0.1,
	}
	e, _, err := effect.New("palette", bdc.PixelCount, ec)
	if err != nil {
		t.Error(err)
	}

	p := make(color.Pixels, bdc.PixelCount)

	t.Run(fmt.Sprintf("%d pixels", bdc.PixelCount), func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			e.Render(p)
			err = d.Send(p)
			if err != nil {
				t.Error(err)
			}
		}
	})
}
