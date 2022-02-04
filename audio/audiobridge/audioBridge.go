package audiobridge

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"ledfx/audio"
	log "ledfx/logger"
)

// NewBridge initializes a new bridge between a source and destination audio device.
func NewBridge(bufferCallback func(buf audio.Buffer)) (br *Bridge, err error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("error initializing PortAudio: %w", err)
	}

	cbHandler := &callbackWrapper{
		callback: bufferCallback,
	}

	br = &Bridge{
		bufferCallback: bufferCallback,
		byteWriter:     &audio.ByteWriter{},
		intWriter:      cbHandler,
		inputType:      inputType(-1), // -1 signifies undefined
		done:           make(chan bool),
	}
	return br, nil
}

func (cbw *callbackWrapper) Write(b audio.Buffer) (int, error) {
	cbw.callback(b)
	return len(b), nil
}

// Stop stops the bridge. Any further references to 'br *Bridge'
// may cause a runtime panic.
func (br *Bridge) Stop() {
	defer func() {
		go func() {
			br.done <- true
		}()
	}()
	if br.airplay != nil {
		log.Logger.WithField("category", "Audio Bridge").Warnf("Stopping AirPlay handler...")
		br.airplay.Stop()
	}

	if br.local != nil {
		log.Logger.WithField("category", "Audio Bridge").Warnf("Stopping local audio handler...")
		br.local.Stop()
	}

	log.Logger.WithField("category", "Audio Bridge").Warnf("Terminating PortAudio...")
	_ = portaudio.Terminate()
}

// Wait waits for the bridge to finish.
func (br *Bridge) Wait() {
	<-br.done
}
