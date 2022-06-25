package audiobridge

import (
	"fmt"
	"ledfx/audio/audiobridge/capture"
	"ledfx/audio/audiobridge/playback"
	"ledfx/config"
	log "ledfx/logger"
)

type LocalHandler struct {
	playback playback.Handler
	capture  *capture.Handler
}

func newLocalHandler() *LocalHandler {
	return &LocalHandler{}
}

func (br *Bridge) StartLocalInput(id string) (err error) {
	if br.inputType != -1 {
		br.closeInput()
	}

	br.inputType = inputTypeLocal

	if br.local == nil {
		br.local = newLocalHandler()
	}

	if br.local.capture != nil {
		br.local.capture.Quit()
	}

	log.Logger.WithField("context", "Local Capture Init").Infof("Initializing new capture handler...")
	if br.local.capture, err = capture.NewHandler(id, br.byteWriter); err != nil {
		return fmt.Errorf("error initializing new capture handler: %w", err)
	}
	config.SetLocalInput(id)

	return nil
}

func (br *Bridge) AddLocalOutput() (err error) {
	if br.local == nil {
		br.local = newLocalHandler()
	}

	if br.local.playback != nil {
		log.Logger.WithField("context", "Local Playback Init").Warn("Local playback already exists! Resetting playback handler...")
		id := br.local.playback.Identifier()
		br.local.playback.Quit()
		if err := br.byteWriter.RemoveWriter(id); err != nil {
			return fmt.Errorf("error removing writer: %w", err)
		}
	}

	log.Logger.WithField("context", "Local Playback Init").Info("Initializing new playback handler...")
	if br.local.playback, err = playback.NewHandler(); err != nil {
		return fmt.Errorf("error initializing new playback handler: %w", err)
	}

	log.Logger.WithField("context", "Local Playback Init").Debug("Wiring local playback output to existing source...")

	if err := br.wireLocalOutput(br.local.playback); err != nil {
		return fmt.Errorf("error wiring local output: %w", err)
	}

	return nil
}

func (lh *LocalHandler) Stop() {
	if lh.capture != nil {
		log.Logger.WithField("context", "Local Audio UnixHandler").Warnf("Stopping capture handler...")
		lh.capture.Quit()
	}
	if lh.playback != nil {
		log.Logger.WithField("context", "Local Audio UnixHandler").Warnf("Stopping playback handler...")
		lh.playback.Quit()
	}
}
