package audiobridge

import (
	"github.com/hajimehoshi/oto"
	"io"
	"ledfx/audio"
	"ledfx/config"
	log "ledfx/logger"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAudioBridge_Darwin(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {})
	if err != nil {
		t.Fatalf("Error creating new bridge: %v\n", err)
	}

	devices, err := audio.GetAudioDevices()
	if err != nil {
		t.Fatalf("Error getting available audio devices: %v\n", err)
	}

	var inputDev *config.AudioDevice
	for i := range devices {
		if strings.Contains(strings.ToLower(devices[i].Name), "mic") {
			inputDev = &devices[i]
			log.Logger.WithField("category", "AudioBridge Test").Infof("Input Device: %s\n", inputDev.Name)
		}
	}

	if inputDev == nil {
		t.Fatalf("Could not find input audio device containing string 'mic'\n")
	}

	if err := br.StartLocalInput(*inputDev); err != nil {
		t.Fatalf("Error starting local input: %v\n", err)
	}

	fi, err := os.Create("./captured.out")
	if err != nil {
		t.Fatalf("Error creating output file: %v\n", err)
	}
	defer func() {
		fi.Close()
		os.Remove("./captured.out")
	}()

	if err := br.AddOutputWriter(fi); err != nil {
		t.Fatalf("Error adding output writer: %v\n", err)
	}

	log.Logger.WithField("category", "AudioBridge Test").Infof("Make some noise!")
	time.Sleep(10 * time.Second)
	br.Stop()

	// Rewind the file
	if _, err := fi.Seek(0, 0); err != nil {
		t.Fatalf("Error rewinding file: %v\n", err)
	}

	otoCtx, err := oto.NewContext(int(inputDev.SampleRate), inputDev.Channels, 2, 100)
	if err != nil {
		t.Fatalf("Error initializing new OTO context: %v\n", err)
	}

	player := otoCtx.NewPlayer()
	defer player.Close()
	log.Logger.WithField("category", "AudioBridge Test").Infof("Audio capture finished. Playing back audio...")
	io.Copy(player, fi)
	log.Logger.WithField("category", "AudioBridge Test").Infof("Did you hear yourself?")
}
