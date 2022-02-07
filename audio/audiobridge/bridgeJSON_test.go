package audiobridge

import (
	"fmt"
	"ledfx/audio"
	"ledfx/config"
	log "ledfx/logger"
	"strings"
	"testing"
)

func TestBridgeMic2LocalJSON(t *testing.T) {
	br, err := NewBridge(func(buf audio.Buffer) {
		// No audio buffer callback because we aren't processing it into blinky lights
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

	wrapper := br.NewJSONWrapper()

	// BEGIN INPUT CONFIG
	inConf := LocalInputJSON{
		AudioDevice: inputDev,
		Verbose:     true,
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
	outConf := LocalOutputJSON{
		Verbose: true,
	}
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
		// No audio buffer callback because we aren't processing it into blinky lights
	})
	if err != nil {
		t.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop()

	wrapper := br.NewJSONWrapper()

	// BEGIN INPUT CONFIG
	inConf := AirPlayInputJSON{
		Name:    "AirPlay2Local",
		Port:    7000,
		Verbose: false,
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
	outConf := LocalOutputJSON{
		Verbose: true,
	}
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
