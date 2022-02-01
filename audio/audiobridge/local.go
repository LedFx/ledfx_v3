package audiobridge

import (
	"fmt"
	"github.com/dustin/go-broadcast"
	"ledfx/audio"
	"ledfx/audio/audiobridge/capture"
	"ledfx/audio/audiobridge/playback"
	"ledfx/config"
	log "ledfx/logger"
)

type LocalHandler struct {
	playback   *playback.Handler
	capture    *capture.Handler
	hermes     broadcast.Broadcaster // Hermes is a messenger for audio buffers.
	hermesChan chan interface{}
}

func newLocalHandler(hermes broadcast.Broadcaster) *LocalHandler {
	lh := &LocalHandler{
		hermes:     hermes,
		hermesChan: make(chan interface{}),
	}
	lh.hermes.Register(lh.hermesChan)
	return lh
}

func (br *Bridge) StartLocalInput(audioDevice config.AudioDevice) (err error) {
	if br.inputType != -1 {
		return fmt.Errorf("an input source has already been defined for this bridge")
	}

	br.inputType = inputTypeLocal

	if br.local == nil {
		br.local = newLocalHandler(br.hermes)
	}

	if br.local.capture == nil {
		if br.local.capture, err = capture.NewHandler(audioDevice, br.local.hermes); err != nil {
			return fmt.Errorf("error initializing new capture handler: %w", err)
		}
	}

	go func() {
		for captured := range br.local.hermesChan {
			br.bufferCallback(captured.(audio.Buffer))
		}
	}()

	return nil
}

func (br *Bridge) AddLocalOutput(audioDevice config.AudioDevice) (err error) {
	if br.local == nil {
		br.local = newLocalHandler(br.hermes)
	}

	if br.local.playback == nil {
		if br.local.playback, err = playback.NewHandler(audioDevice, br.local.hermes); err != nil {
			return fmt.Errorf("error initializing new playback handler: %w", err)
		}
	}

	return nil
}

func (lh *LocalHandler) Stop() {
	if lh.capture != nil {
		log.Logger.WithField("category", "Local Audio Handler").Warnf("Stopping capture handler...")
		lh.capture.Quit()
	}
	if lh.playback != nil {
		log.Logger.WithField("category", "Local Audio Handler").Warnf("Stopping playback handler...")
		lh.playback.Quit()
	}
}
