package audiobridge

import (
	"fmt"
	"ledfx/audio"
	log "ledfx/logger"
	"strings"
	"testing"
)

func TestBridgeMic2LocalJSON(t *testing.T) {
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

	var deviceId string
	for i := range devices {
		if strings.Contains(strings.ToLower(devices[i].Name), "mic") {
			deviceId = devices[i].Id
			log.Logger.WithField("context", "Mic2Speaker").Infof("Mic Device: %s", devices[i].Name)
			log.Logger.WithField("context", "Mic2Speaker").Infof("Mic Input Channels: %d", devices[i].ChannelsIn)
			log.Logger.WithField("context", "Mic2Speaker").Infof("Mic Sample Rate: %f", devices[i].SampleRate)
			break
		}
	}

	if deviceId == "" {
		t.Fatalf("Could not find input audio device containing string 'mic'\n")
	}

	wrapper := br.JSONWrapper()

	// BEGIN INPUT CONFIG
	inConf := LocalInputJSON{
		DeviceID: deviceId,
	}
	inConfBytes, err := inConf.AsJSON()
	if err != nil {
		t.Fatalf("Error marshalling input config JSON: %v\n", err)
	}
	if err := wrapper.StartLocalInput(inConfBytes); err != nil {
		t.Fatalf("Error starting local input: %v\n", err)
	}
	// END INPUT CONFIG

	// BEGIN OUTPUT CONFIG
	outConf := LocalOutputJSON{}
	outConfBytes, err := outConf.AsJSON()
	if err != nil {
		t.Fatalf("Error marshalling output config JSON: %v\n", err)
	}
	if err := wrapper.AddLocalOutput(outConfBytes); err != nil {
		t.Fatalf("Error adding local output: %v\n", err)
	}
	// END OUTPUT CONFIG

	fmt.Printf("\n##--####--####--####--####--####--####--####--####--####--####--##\n")
	fmt.Printf("Output JSON: [%s]\n", string(outConfBytes))
	fmt.Printf("Input JSON: [%s]\n", string(inConfBytes))
	fmt.Printf("##--####--####--####--####--####--####--####--####--####--####--##\n\n")

	br.Wait()
}

func TestBridgeAirPlay2LocalJSON(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {
		// No audio buffer Callback because we aren't processing it into blinky lights
	})
	if err != nil {
		t.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	wrapper := br.JSONWrapper()

	// BEGIN INPUT CONFIG
	inConf := AirPlayInputJSON{
		Name: "AirPlay2Local",
		Port: 7000,
	}
	inConfBytes, err := inConf.AsJSON()
	if err != nil {
		t.Fatalf("Error marshalling input config JSON: %v\n", err)
	}
	if err := wrapper.StartAirPlayInput(inConfBytes); err != nil {
		t.Fatalf("Error starting local input: %v\n", err)
	}
	// END INPUT CONFIG

	// BEGIN OUTPUT CONFIG
	outConf := LocalOutputJSON{}
	outConfBytes, err := outConf.AsJSON()
	if err != nil {
		t.Fatalf("Error marshalling output config JSON: %v\n", err)
	}
	if err := wrapper.AddLocalOutput(outConfBytes); err != nil {
		t.Fatalf("Error adding local output: %v\n", err)
	}
	// END OUTPUT CONFIG

	fmt.Printf("\n##--####--####--####--####--####--####--####--####--####--####--##\n")
	fmt.Printf("Output JSON: [%s]\n", string(outConfBytes))
	fmt.Printf("Input JSON: [%s]\n", string(inConfBytes))
	fmt.Printf("##--####--####--####--####--####--####--####--####--####--####--##\n\n")

	br.Wait()
}

func TestBridgeYouTube2LocalJSON(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {
		// No audio buffer Callback because we aren't processing it into blinky lights
	})
	if err != nil {
		t.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	wrapper := br.JSONWrapper()

	// BEGIN INPUT CONFIG
	inConf := YouTubeInputJSON{}
	inConfBytes, err := inConf.AsJSON()
	if err != nil {
		t.Fatalf("Error marshalling input config JSON: %v\n", err)
	}

	if err := wrapper.StartYouTubeInput(inConfBytes); err != nil {
		t.Fatalf("Error starting YouTubeSet input: %v\n", err)
	}
	// END INPUT CONFIG

	// BEGIN OUTPUT CONFIG
	outConf := LocalOutputJSON{}
	outConfBytes, err := outConf.AsJSON()
	if err != nil {
		t.Fatalf("Error marshalling output config JSON: %v\n", err)
	}
	if err := wrapper.AddLocalOutput(outConfBytes); err != nil {
		t.Fatalf("Error adding local output: %v\n", err)
	}
	// END OUTPUT CONFIG

	fmt.Printf("\n##--####--####--####--####--####--####--####--####--####--####--##\n")
	fmt.Printf("Output JSON: [%s]\n", string(outConfBytes))
	fmt.Printf("Input JSON: [%s]\n", string(inConfBytes))
	fmt.Printf("##--####--####--####--####--####--####--####--####--####--####--##\n\n")

	// BEGIN DOWNLOAD CONFIG
	ctlConf := YouTubeCTLJSON{
		Action: YouTubeActionDownload,
		URL:    "https://www.youtube.com/watch?v=ML1U9PDcFWk",
	}
	ctlConfBytes, err := ctlConf.AsJSON()
	if err != nil {
		t.Fatalf("Error marshalling download control config as JSON: %v\n", err)
	}
	if _, err = wrapper.CTL().YouTubeSet(ctlConfBytes); err != nil {
		t.Fatalf("Error downloading video: %v\n", err)
	}
	// END DOWNLOAD CONFIG

	// BEGIN PLAY CONFIG
	ctlConf = YouTubeCTLJSON{
		Action: YouTubeActionPlay,
	}
	ctlConfBytes, err = ctlConf.AsJSON()
	if err != nil {
		t.Fatalf("Error marshalling play control config as JSON: %v\n", err)
	}
	if _, err = wrapper.CTL().YouTubeSet(ctlConfBytes); err != nil {
		t.Fatalf("Error playing audio from video: %v\n", err)
	}
	// END PLAY CONFIG

	br.Wait()
}
