package audiobridge

import (
	"ledfx/audio"
	"ledfx/config"
	log "ledfx/logger"
	"strings"
	"testing"
	"time"
)

func TestBridgeMic2Speaker(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {})
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
			numChannels = inputDev.Channels
			sampleRate = int(inputDev.SampleRate)
			log.Logger.WithField("category", "Mic2Speaker").Infof("Mic Device: %s", inputDev.Name)
			log.Logger.WithField("category", "Mic2Speaker").Infof("Mic Input Channels: %d", numChannels)
			log.Logger.WithField("category", "Mic2Speaker").Infof("Mic Sample Rate: %d", sampleRate)
			break
		}
	}

	if inputDev == nil {
		t.Fatalf("Could not find input audio device containing string 'mic'\n")
	}

	if err := br.StartLocalInput(*inputDev); err != nil {
		t.Fatalf("Error starting local input: %v\n", err)
	}

	if err := br.AddLocalOutput(); err != nil {
		t.Fatalf("Error adding local output: %v\n", err)
	}

	time.Sleep(10 * time.Second)
}
