package integrationTests

import (
	"io"
	"ledfx/integrations/audiobridge"
	"testing"
)

func TestAudioBridge(t *testing.T) {
	srcConf := audiobridge.EndpointConfig{
		Type:    audiobridge.DeviceTypeAirPlay,
		Name:    "LedFX-Input-Test",
		Verbose: true,
	}
	dstConf := audiobridge.EndpointConfig{
		Type:    audiobridge.DeviceTypeAirPlay,
		Name:    "(?i)Output$",
		Verbose: true,
	}
	bridge, err := audiobridge.NewBridge(srcConf, dstConf, io.Discard)
	if err != nil {
		t.Fatalf("error creating new audio bridge: %v\n", err)
	}
	bridge.Wait()
}
