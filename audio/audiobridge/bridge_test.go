package audiobridge

import (
	"bufio"
	"io"
	"ledfx/audio"
	"ledfx/config"
	log "ledfx/logger"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestBridgeMic2Local(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {
		// No audio buffer Callback because we aren't processing it into blinky lights
	})
	if err != nil {
		t.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	devices, err := audio.GetAudioDevices()
	if err != nil {
		t.Fatalf("Error getting available audio devices: %v\n", err)
	}

	var inputDev *config.AudioDevice
	for i := range devices {
		if strings.Contains(strings.ToLower(devices[i].Name), "mic") {
			inputDev = &devices[i]
			log.Logger.WithField("category", "Mic2Speaker").Infof("Mic Device: %s", inputDev.Name)
			log.Logger.WithField("category", "Mic2Speaker").Infof("Mic Input Channels: %d", inputDev.Channels)
			log.Logger.WithField("category", "Mic2Speaker").Infof("Mic Sample Rate: %f", inputDev.SampleRate)
			break
		}
	}

	if inputDev == nil {
		t.Fatalf("Could not find input audio device containing string 'mic'\n")
	}

	if err := br.StartLocalInput(*inputDev, true); err != nil {
		t.Fatalf("Error starting local input: %v\n", err)
	}

	if err := br.AddLocalOutput(true); err != nil {
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

	if err := br.StartAirPlayInput("LedFX-AirPlay", 7000, false); err != nil {
		t.Fatalf("Error initializing AirPlay input: %v\n", err)
	}

	if err := br.AddLocalOutput(true); err != nil {
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

	if err := br.StartAirPlayInput("LedFX-Test", 7000, false); err != nil {
		t.Fatalf("Error initializing AirPlay input: %v\n", err)
	}

	if err := br.AddAirPlayOutput("LedFX-AirPlay", AirPlaySearchByName, true); err != nil {
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

	if err := br.StartAirPlayInput("LedFX-Test-Input", 7000, false); err != nil {
		t.Fatalf("Error initializing AirPlay input: %v\n", err)
	}

	if err := br.AddAirPlayOutput("LedFX-AirPlay", AirPlaySearchByName, true); err != nil {
		t.Fatalf("Error initializing AirPlay output: %v\n", err)
	}

	if err := br.AddLocalOutput(true); err != nil {
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

	if err := br.StartAirPlayInput("LedFX-Test", 7000, false); err != nil {
		t.Fatalf("Error initializing AirPlay input: %v\n", err)
	}

	if err := br.AddAirPlayOutput("LedFX-AirPlay", AirPlaySearchByName, true); err != nil {
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

	devices, err := audio.GetAudioDevices()
	if err != nil {
		t.Fatalf("Error getting available audio devices: %v\n", err)
	}

	var inputDev *config.AudioDevice
	for i := range devices {
		if strings.Contains(strings.ToLower(devices[i].Name), "mic") {
			inputDev = &devices[i]
			log.Logger.WithField("category", "Mic2Speaker").Infof("Mic Device: %s", inputDev.Name)
			log.Logger.WithField("category", "Mic2Speaker").Infof("Mic Input Channels: %d", inputDev.Channels)
			log.Logger.WithField("category", "Mic2Speaker").Infof("Mic Sample Rate: %f", inputDev.SampleRate)
			break
		}
	}

	if inputDev == nil {
		t.Fatalf("Could not find input audio device containing string 'mic'\n")
	}

	if err := br.StartLocalInput(*inputDev, true); err != nil {
		t.Fatalf("Error starting local input: %v\n", err)
	}

	if err := br.AddAirPlayOutput("LedFX-AirPlay", AirPlaySearchByName, true); err != nil {
		t.Fatalf("Error adding AirPlay output: %v\n", err)
	}

	br.Wait()
}

func TestBridgeYoutube2Local(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {
		// No audio buffer Callback because we aren't processing it into blinky lights.
	})
	if err != nil {
		t.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	if err := br.StartYoutubeInput(true); err != nil {
		t.Fatalf("Error starting YouTube input: %v\n", err)
	}

	if err := br.AddLocalOutput(true); err != nil {
		t.Fatalf("Error starting local output: %v\n", err)
	}

	pp, err := br.Controller().YouTube().PlayPlaylist("https://youtube.com/playlist?list=PLcncP1HGs_p0VaCVjUPyrPiRSbQS8H8W-")
	if err != nil {
		t.Fatalf("Error playing YouTube playlist: %v\n", err)
	}
	defer pp.Stop()

	go func() {
		for {
			if err := pp.Next(); err != nil {
				log.Logger.Warnf("Error playing: %v", err)
			}
			time.Sleep(5 * time.Second)
		}

	}()

	br.Wait()
}
