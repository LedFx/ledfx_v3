package controller

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/LedFx/ledfx/pkg/audio"
	"github.com/LedFx/ledfx/pkg/audio/audiobridge"
	"github.com/LedFx/ledfx/pkg/color"
	"github.com/LedFx/ledfx/pkg/device"
	"github.com/LedFx/ledfx/pkg/effect"
)

func TestController(t *testing.T) {

	bdc := map[string]interface{}{
		"pixel_count": 64,
		"name":        "Spotlight",
	}
	udpc := map[string]interface{}{
		"ip":       "192.168.0.104",
		"port":     21324,
		"protocol": "DRGB",
		"timeout":  60,
	}
	d, _, err := device.New("", "udp", bdc, udpc)
	if err != nil {
		t.Error(err)
	}
	err = d.Connect()
	if err != nil {
		t.Error(err)
	}
	ec := map[string]interface{}{
		"intensity":      0.5,
		"brightness":     1,
		"saturation":     1,
		"palette":        "RGB",
		"blur":           0.9,
		"decay":          0.1,
		"flip":           false,
		"mirror":         true,
		"bkg_brightness": 1,
		"bkg_color":      "#000000",
		"hue_shift":      0.00001,
	}
	e, _, err := effect.New("", "energy", bdc["pixel_count"].(int), ec)
	if err != nil {
		t.Error(err)
	}

	p := make(color.Pixels, bdc["pixel_count"].(int))

	br, err := audiobridge.NewBridge(audio.Analyzer.BufferCallback)
	if err != nil {
		log.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	if err := br.StartLocalInput(""); err != nil { // get from config
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

func BenchmarkController(t *testing.B) {

	bdc := map[string]interface{}{
		"pixel_count": 64,
		"name":        "Spotlight",
	}
	udpc := map[string]interface{}{
		"ip":       "192.168.0.104",
		"port":     21324,
		"protocol": "DRGB",
		"timeout":  60,
	}
	d, _, err := device.New("", "udp", bdc, udpc)
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
	e, _, err := effect.New("", "palette", bdc["pixel_count"].(int), ec)
	if err != nil {
		t.Error(err)
	}

	p := make(color.Pixels, bdc["pixel_count"].(int))

	t.Run(fmt.Sprintf("%d pixels", bdc["pixel_count"].(int)), func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			e.Render(p)
			err = d.Send(p)
			if err != nil {
				t.Error(err)
			}
		}
	})
}
