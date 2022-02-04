package airplay2

import (
	"github.com/dustin/go-broadcast"
	"ledfx/audio/audiobridge/playback"
	log "ledfx/logger"
	"os"
	"os/signal"
	"testing"
)

func TestAirPlayServer(t *testing.T) {
	hermes := broadcast.NewBroadcaster(60)

	sv := NewServer(Config{
		AdvertisementName: "LedFX-AirPlay",
		VerboseLogging:    false,
		Port:              7000,
	}, hermes)

	handler, err := playback.NewHandler(hermes)
	if err != nil {
		t.Fatalf("Error initializing playback handler: %v\n", err)
	}
	defer handler.Quit()

	if err := sv.Start(); err != nil {
		t.Fatalf("Error starting AirPlay server: %v\n", err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		log.Logger.WithField("category", "TEST: AirPlay Server").Warnf("Stopping AirPlay server...")
		sv.Stop()
	}()

	sv.Wait()
}
