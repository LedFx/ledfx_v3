package audiobridge

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/LedFx/ledfx/pkg/audio"
	log "github.com/LedFx/ledfx/pkg/logger"
	"github.com/gen2brain/malgo"
)

func TestBridgeMic2Local(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {
		// No audio buffer Callback because we aren't processing it into blinky lights
	})
	if err != nil {
		t.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	devices, err := audio.Context.Devices(malgo.Capture)
	if err != nil {
		t.Fatalf("Error getting available audio devices: %v\n", err)
	}

	var deviceId string
	for i := range devices {
		if strings.Contains(strings.ToLower(devices[i].Name()), "mic") {
			deviceId = devices[i].ID.String()
			log.Logger.WithField("context", "Mic2Speaker").Infof("Mic Device: %s", devices[i].Name)
			log.Logger.WithField("context", "Mic2Speaker").Infof("Mic Input Channels: %d - %d", devices[i].MinChannels, devices[i].MaxChannels)
			log.Logger.WithField("context", "Mic2Speaker").Infof("Mic Sample Rate: %d - %d", devices[i].MinSampleRate, devices[i].MaxSampleRate)
			break
		}
	}

	if deviceId == "" {
		t.Fatalf("Could not find input audio device containing string 'mic'\n")
	}

	if err := br.StartLocalInput(deviceId); err != nil {
		t.Fatalf("Error starting local input: %v\n", err)
	}

	if err := br.AddLocalOutput(); err != nil {
		t.Fatalf("Error adding local output: %v\n", err)
	}

	br.Wait()
}

func TestBridgeAirplay2Local(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {
		// No audio buffer Callback because we aren't processing it into blinky lights.
	})
	if err != nil {
		t.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	if err := br.StartAirPlayInput("LedFX-AirPlay", 7000); err != nil {
		t.Fatalf("Error initializing AirPlay input: %v\n", err)
	}

	if err := br.AddLocalOutput(); err != nil {
		t.Fatalf("Error initializing local output: %v\n", err)
	}

	br.Wait()
}

func TestBridgeAirPlay2AirPlay(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {
		// No audio buffer Callback because we aren't processing it into blinky lights.
	})
	if err != nil {
		t.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	if err := br.StartAirPlayInput("LedFX-Test", 7000); err != nil {
		t.Fatalf("Error initializing AirPlay input: %v\n", err)
	}

	if err := br.AddAirPlayOutput("LedFX-AirPlay", AirPlaySearchByName); err != nil {
		t.Fatalf("Error initializing AirPlay output: %v\n", err)
	}

	br.Wait()
}

func TestBridgeAirPlay2AirPlayAndLocal(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {
		// No audio buffer Callback because we aren't processing it into blinky lights.
	})
	if err != nil {
		t.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	if err := br.StartAirPlayInput("LedFX-Test-Input", 7000); err != nil {
		t.Fatalf("Error initializing AirPlay input: %v\n", err)
	}

	if err := br.AddAirPlayOutput("LedFX-AirPlay", AirPlaySearchByName); err != nil {
		t.Fatalf("Error initializing AirPlay output: %v\n", err)
	}

	if err := br.AddLocalOutput(); err != nil {
		t.Fatalf("Error adding local output")
	}

	br.Wait()
}

func TestBridgeAirPlay2AirPlayAsyncWrite(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {
		// No audio buffer Callback because we aren't processing it into blinky lights.
	})
	if err != nil {
		t.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	if err := br.StartAirPlayInput("LedFX-Test", 7000); err != nil {
		t.Fatalf("Error initializing AirPlay input: %v\n", err)
	}

	if err := br.AddAirPlayOutput("LedFX-AirPlay", AirPlaySearchByName); err != nil {
		t.Fatalf("Error initializing AirPlay output: %v\n", err)
	}

	// Add a bunch of writers, so we can listen for delay issues.
	for i := 0; i < 10; i++ {
		if err := br.AddOutputWriter(bufio.NewWriter(io.Discard), strconv.Itoa(i)); err != nil {
			t.Fatalf("Error adding output writer: %v\n", err)
		}
	}

	br.Wait()
}

func TestBridgeMic2AirPlay(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {
		// No audio buffer Callback because we aren't processing it into blinky lights.
	})
	if err != nil {
		t.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	devices, err := audio.Context.Devices(malgo.Capture)
	if err != nil {
		t.Fatalf("Error getting available audio devices: %v\n", err)
	}

	var deviceId string
	for i := range devices {
		if strings.Contains(strings.ToLower(devices[i].Name()), "mic") {
			deviceId = devices[i].ID.String()
			log.Logger.WithField("context", "Mic2Speaker").Infof("Mic Device: %s", devices[i].Name)
			log.Logger.WithField("context", "Mic2Speaker").Infof("Mic Input Channels: %d - %d", devices[i].MinChannels, devices[i].MaxChannels)
			log.Logger.WithField("context", "Mic2Speaker").Infof("Mic Sample Rate: %d - %d", devices[i].MinSampleRate, devices[i].MaxSampleRate)
			break
		}
	}

	if deviceId == "" {
		t.Fatalf("Could not find input audio device containing string 'mic'\n")
	}

	if err := br.AddAirPlayOutput("LedFX-AirPlay", AirPlaySearchByName); err != nil {
		t.Fatalf("Error adding AirPlay output: %v\n", err)
	}

	br.Wait()
}
