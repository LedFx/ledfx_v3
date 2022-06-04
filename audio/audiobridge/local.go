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
	verbose  bool
}

func newLocalHandler(verbose bool) *LocalHandler {
	return &LocalHandler{verbose: verbose}
}

func (br *Bridge) StartLocalInput(audioDevice config.AudioDevice, verbose bool) (err error) {
	if br.inputType != -1 {
		br.closeInput()
	}

	br.inputType = inputTypeLocal

	if br.local == nil {
		br.local = newLocalHandler(verbose)
	}

	if br.local.capture != nil {
		br.local.capture.Quit()
	}

	log.Logger.WithField("category", "Local Capture Init").Infof("Initializing new capture handler...")
	if br.local.capture, err = capture.NewHandler(audioDevice, br.byteWriter, verbose); err != nil {
		return fmt.Errorf("error initializing new capture handler: %w", err)
	}

	return nil
}

func (br *Bridge) AddLocalOutput(verbose bool) (err error) {
	if br.local == nil {
		br.local = newLocalHandler(verbose)
	}

	if br.local.playback.Device() != "" {
		if verbose {
			log.Logger.WithField("category", "Local Playback Init").Warnln("Local playback already exists! Resetting playback handler...")
		}
		id := br.local.playback.Identifier()
		br.local.playback.Quit()
		if err := br.byteWriter.RemoveWriter(id); err != nil {
			return fmt.Errorf("error removing writer: %w", err)
		}
	}

	log.Logger.WithField("category", "Local Playback Init").Infof("Initializing new playback handler...")
	if br.local.playback, err = playback.NewHandler(verbose); err != nil {
		return fmt.Errorf("error initializing new playback handler: %w", err)
	}

	if verbose {
		log.Logger.WithField("category", "Local Playback Init").Infof("Wiring local playback output to existing source...")
	}
	if err := br.wireLocalOutput(br.local.playback); err != nil {
		return fmt.Errorf("error wiring local output: %w", err)
	}

	return nil
}

func (lh *LocalHandler) Stop() {
	if lh.capture != nil {
		log.Logger.WithField("category", "Local Audio UnixHandler").Warnf("Stopping capture handler...")
		lh.capture.Quit()
	}
	if lh.playback.Device() != "" {
		log.Logger.WithField("category", "Local Audio UnixHandler").Warnf("Stopping playback handler...")
		lh.playback.Quit()
	}
}
