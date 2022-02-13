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

	cbHandler := &CallbackWrapper{
		Callback: bufferCallback,
	}

	br = &Bridge{
		bufferCallback: bufferCallback,
		byteWriter:     audio.NewAsyncMultiWriter(),
		intWriter:      cbHandler,
		inputType:      inputType(-1), // -1 signifies undefined
		done:           make(chan bool),
	}
	br.ctl = br.newController()
	return br, nil
}

func (cbw *CallbackWrapper) Write(b audio.Buffer) (int, error) {
	cbw.Callback(b)
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

func (br *Bridge) closeInput() {
	switch br.inputType {
	case inputTypeAirPlayServer:
		br.airplay.server.Stop()
	case inputTypeLocal:
		br.local.capture.Quit()
	case inputTypeYoutube:
		br.youtube.handler.Quit()
	}
}

// Wait waits for the bridge to finish.
func (br *Bridge) Wait() {
	<-br.done
}
