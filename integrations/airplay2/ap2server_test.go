package airplay2

import (
	"github.com/hajimehoshi/oto"
	"ledfx/config"
	log "ledfx/logger"
	"os"
	"testing"
	"time"
)

func TestAirPlayServer(t *testing.T) {
	if _, err := log.Init(config.Config{
		Verbose: true,
	}); err != nil {
		t.Fatalf("Error initializing logger: %v\n", err)
	}

	server := NewServer(Config{
		AdvertisementName: "AirPlay2-TestServer",
		VerboseLogging:    false,
	})

	otoCtx, err := oto.NewContext(44100, 2, 2, 10000)
	if err != nil {
		t.Fatalf("Error creating new oto context: %v\n", err)
	}

	outputAudio := otoCtx.NewPlayer()
	defer outputAudio.Close()

	server.AddOutput(outputAudio)
	if err := server.Start(); err != nil {
		t.Fatalf("Error starting AirPlay2 server: %v\n", err)
	}

	go func() {
		time.Sleep(1 * time.Minute)
		if err := server.Stop(); err != nil {
			t.Errorf("Error stopping server: %v\n", err)
			os.Exit(1)
		}
	}()

	server.Wait()
}
