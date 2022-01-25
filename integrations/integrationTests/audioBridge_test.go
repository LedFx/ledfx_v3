package integrationTests

import (
	"io"
	"ledfx/integrations/audiobridge"
	"testing"
)

func TestAudioBridge(t *testing.T) {
	// srcConf is the config for the endpoint from which
	// audio will be ingested, processed, converted, and
	// redistributed.
	srcConf := audiobridge.EndpointConfig{
		Type:    audiobridge.DeviceTypeAirPlay, // Spin up an AirPlay server
		Name:    "LedFX-Input-Test",            // It will be advertised as "LedFX-Input-Test"
		Verbose: true,                          // Enable verbosity because this is a test package.
	}

	dstConf := audiobridge.EndpointConfig{
		Type: audiobridge.DeviceTypeAirPlay, // Connect to an airplay server

		// We will connect to any AirPlay server that contains the string "Output" (case-insensitive)
		Name: "(?i)Output$", // Regular expressions are required here.

		Verbose: true, // Enable verbosity because this is a test package.
	}

	// Initialize the bridge. This will start everything, too.
	bridge, err := audiobridge.NewBridge(srcConf, dstConf, io.Discard)
	if err != nil {
		t.Fatalf("error creating new audio bridge: %v\n", err)
	}

	// Wait until the bridge loop stops. This can be called with Stop()
	bridge.Wait()

}
