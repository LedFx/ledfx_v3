package audiobridge

import (
	"fmt"
	"ledfx/audio/audiobridge/capture"
	"ledfx/audio/audiobridge/playback"
	"ledfx/config"
	log "ledfx/logger"
)

type LocalHandler struct {
	playback *playback.Handler
	capture  *capture.Handler
}

func newLocalHandler() *LocalHandler {
	return &LocalHandler{}
}

func (br *Bridge) StartLocalInput(audioDevice config.AudioDevice) (err error) {
	if br.inputType != -1 {
		return fmt.Errorf("an input source has already been defined for this bridge")
	}

	br.inputType = inputTypeLocal

	if br.local == nil {
		br.local = newLocalHandler()
	}

	if br.local.capture == nil {
		if br.local.capture, err = capture.NewHandler(audioDevice, br.intWriter, br.byteWriter); err != nil {
			return fmt.Errorf("error initializing new capture handler: %w", err)
		}
	}

	return nil
}

func (br *Bridge) AddLocalOutput() (err error) {
	if br.local == nil {
		br.local = newLocalHandler()
	}

	if br.local.playback == nil {
		if br.local.playback, err = playback.NewHandler(); err != nil {
			return fmt.Errorf("error initializing new playback handler: %w", err)
		}
	}

	br.wireLocalOutput(br.local.playback)

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
