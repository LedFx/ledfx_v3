package audiobridge

import (
	"bytes"
	"github.com/carterpeel/oto/v2"
	"io/ioutil"
	"ledfx/audio"
	"ledfx/config"
	log "ledfx/logger"
	"os"
	"strings"
	"testing"
	"time"
)

func initFile() {
	var err error
	if fi, err = os.OpenFile("./captured.out", os.O_CREATE|os.O_RDWR, 0777); err != nil {
		log.Logger.Fatalf("Error creating output file: %v\n", err)
	}
	fi.Seek(0, 0)
}
func rewindFile() {
	fi.Seek(0, 0)
}

var (
	fi          *os.File
	numChannels int
	sampleRate  int
)

func TestLoopbackAudioCapture_Darwin(t *testing.T) {
	initFile()

	br, err := NewBridge(func(buf audio.Buffer) {})
	if err != nil {
		t.Fatalf("Error creating new bridge: %v\n", err)
	}
	defer br.Stop()

	devices, err := audio.GetAudioDevices()
	if err != nil {
		t.Fatalf("Error getting available audio devices: %v\n", err)
	}

	var inputDev *config.AudioDevice
	for i := range devices {
		if strings.Contains(strings.ToLower(devices[i].Name), "blackhole") {
			inputDev = &devices[i]
			numChannels = inputDev.Channels
			sampleRate = int(inputDev.SampleRate)
			log.Logger.WithField("category", "Loopback Capture").Infof("Loopback Input Device: %s", inputDev.Name)
			log.Logger.WithField("category", "Loopback Capture").Infof("Loopback Input Channels: %d", numChannels)
			log.Logger.WithField("category", "Loopback Capture").Infof("Loopback Sample Rate: %d", sampleRate)
			break
		}
	}

	if inputDev == nil {
		t.Fatalf("Could not find input audio device containing string 'mic'\n")
	}

	if err := br.StartLocalInput(*inputDev); err != nil {
		t.Fatalf("Error starting local input: %v\n", err)
	}

	if err := br.AddOutputWriter(fi); err != nil {
		t.Fatalf("Error adding output writer: %v\n", err)
	}

	log.Logger.WithField("category", "Loopback Capture").Infof("Play some audio!")
	time.Sleep(20 * time.Second)
	log.Logger.WithField("category", "Loopback Capture").Infof("Wrote captured audio from loopback to './captured.out'")

	t.Cleanup(playbackHandler)
}

func TestMicrophoneAudioCapture_Darwin(t *testing.T) {
	initFile()

	br, err := NewBridge(func(buf audio.Buffer) {})
	if err != nil {
		t.Fatalf("Error creating new bridge: %v\n", err)
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
			log.Logger.WithField("category", "Mic Capture").Infof("Mic Device: %s", inputDev.Name)
			log.Logger.WithField("category", "Mic Capture").Infof("Mic Input Channels: %d", numChannels)
			log.Logger.WithField("category", "Mic Capture").Infof("Mic Sample Rate: %d", sampleRate)
			break
		}
	}

	if inputDev == nil {
		t.Fatalf("Could not find input audio device containing string 'mic'\n")
	}

	if err := br.StartLocalInput(*inputDev); err != nil {
		t.Fatalf("Error starting local input: %v\n", err)
	}

	if err := br.AddOutputWriter(fi); err != nil {
		t.Fatalf("Error adding output writer: %v\n", err)
	}

	log.Logger.WithField("category", "Mic Capture").Infof("Make some noise!")
	time.Sleep(10 * time.Second)
	log.Logger.WithField("category", "Mic Capture").Infof("Wrote captured audio to './captured.out'")

	t.Cleanup(playbackHandler)
}

func playbackHandler() {
	defer func() {
		fi.Close()
		os.Remove("./captured.out")
	}()

	log.Logger.WithField("category", "Capture Playback").Infof("Sample Rate: %d", sampleRate)
	log.Logger.WithField("category", "Capture Playback").Infof("Output Channels: %d", numChannels)
	ctx, ready, err := oto.NewContext(sampleRate, numChannels, 2)
	if err != nil {
		log.Logger.Fatalf("Error initializing new OTO context: %v\n", err)
	}
	<-ready

	rewindFile()
	audioBytes, err := ioutil.ReadAll(fi)
	if err != nil {
		log.Logger.Fatalf("Error reading bytes from file: %v\n", err)
	}

	pl := ctx.NewPlayer(bytes.NewReader(audioBytes))
	defer pl.Close()
	pl.SetVolume(1)
	pl.Play(true)
}
