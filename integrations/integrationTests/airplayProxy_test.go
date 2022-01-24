package integrationTests

import (
	"github.com/hajimehoshi/oto"
	"ledfx/config"
	"ledfx/integrations/airplay2"
	log "ledfx/logger"
	"testing"
)

func init() {
	_, _ = log.Init(config.Config{
		Verbose: true,
	})
}

// Final audio destination (cannot be run on same machine as InputBridge)
func TestAirPlayProxy_OutputServer(t *testing.T) {
	outputServer := airplay2.NewServer(airplay2.Config{
		AdvertisementName: "AirPlay2-Test-OutputServer",
		VerboseLogging:    false,
		Port:              7000,
	})

	otoCtx, err := oto.NewContext(44100, 2, 2, 12000)
	if err != nil {
		t.Fatalf("Error creating audio player context: %v\n", err)
	}
	player := otoCtx.NewPlayer()
	defer player.Close()
	outputServer.AddOutput(player)

	if err := outputServer.Start(); err != nil {
		t.Fatalf("Error starting AirPlay outputServer: %v\n", err)
	}

	outputServer.Wait()
}

// The client bridges the audio between the input server and the output server
func TestAirPlayProxy_InputServer(t *testing.T) {
	inputServer := airplay2.NewServer(airplay2.Config{
		AdvertisementName: "AirPlay2-Test-InputServer",
		VerboseLogging:    false,
		Port:              8093,
	})

	client, err := airplay2.NewClient(airplay2.ClientDiscoveryParameters{
		DeviceName: "OutputServer",
		Verbose:    true,
	})
	if err != nil {
		t.Fatalf("Error connecting to AirPlay server: %v\n", err)
	}
	inputServer.SetClient(client)
	if err := inputServer.Start(); err != nil {
		t.Fatalf("Error starting AirPlay inputServer: %v\n", err)
	}
	inputServer.Wait()
}
