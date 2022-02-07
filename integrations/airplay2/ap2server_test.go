package airplay2

import (
	"ledfx/audio"
	"ledfx/audio/audiobridge/playback"
	log "ledfx/logger"
	"os"
	"os/signal"
	"testing"
)

type intWriterTest struct{}

func (iwt intWriterTest) Write(b audio.Buffer) (n int, err error) {
	// No callback required for this test
	return len(b), nil
}

func TestAirPlayServer(t *testing.T) {
	intWriter := intWriterTest{}
	byteWriter := audio.NewByteWriter()

	handler, err := playback.NewHandler(true)
	if err != nil {
		t.Fatalf("Error initializing playback handler: %v\n", err)
	}
	defer handler.Quit()

	if err := byteWriter.AddWriter(handler, "TEST"); err != nil {
		t.Fatalf("Error adding writer: %v\n", err)
	}

	sv := NewServer(Config{
		AdvertisementName: "LedFX-AirPlay",
		VerboseLogging:    false,
		Port:              7000,
	}, intWriter, byteWriter)

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
