package integrationTests

import (
	"ledfx/integrations/audiobridge"
	"testing"
)

func TestAudioBridge(t *testing.T) {
	srcConf := audiobridge.EndpointConfig{
		Type: audiobridge.DeviceTypeAirPlay,
		Name: "LedFX-AirPlay-Input",
	}
	dstConf := audiobridge.EndpointConfig{
		Type: audiobridge.DeviceTypeAirPlay,
		Name: "LedFX-AirPlay-Output",
	}
	bridge, err := audiobridge.NewBridge(srcConf, dstConf)
	if err != nil {
		t.Fatalf("error creating new audio bridge: %v\n", err)
	}
	bridge.Wait()
}
