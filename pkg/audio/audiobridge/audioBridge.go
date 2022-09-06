package audiobridge

import (
	"fmt"

	"github.com/LedFx/ledfx/pkg/audio"
	"github.com/LedFx/ledfx/pkg/audio/audiobridge/assets"
	log "github.com/LedFx/ledfx/pkg/logger"
)

// NewBridge initializes a new bridge between a source and destination audio device.
func NewBridge(bufferCallback func(buf audio.Buffer)) (br *Bridge, err error) {
	br = &Bridge{
		bufferCallback: bufferCallback,
		byteWriter:     audio.NewAsyncMultiWriter(),
		inputType:      inputType(-1), // -1 signifies undefined
		done:           make(chan bool),
		outputs:        make([]*OutputInfo, 0),
	}

	br.info = &Info{
		br: br,
	}

	if err := br.byteWriter.AddWriter(&CallbackWrapper{
		Callback: bufferCallback,
	}, "CallbackWrapper"); err != nil {
		return nil, fmt.Errorf("error adding callback wrapper to writer: %w", err)
	}

	br.ctl = br.newController()
	return br, nil
}

func (cbw *CallbackWrapper) Write(p []byte) (int, error) {
	cbw.Callback(audio.BytesToAudioBuffer(p))
	return len(p), nil
}

func (br *Bridge) Artwork() []byte {
	if br.Controller().AirPlay().Server() != nil {
		return br.Controller().AirPlay().Server().Artwork()
	}
	return assets.BlankAlbumArt()
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
		log.Logger.WithField("context", "Audio Bridge").Warnf("Stopping AirPlay handler...")
		br.airplay.Stop()
	}

	if br.local != nil {
		log.Logger.WithField("context", "Audio Bridge").Warnf("Stopping local audio handler...")
		br.local.Stop()
	}
}

func (br *Bridge) closeInput() {
	switch br.inputType {
	case inputTypeAirPlayServer:
		if !br.airplay.server.Stopped() {
			br.airplay.server.Stop()
		}
	case inputTypeLocal:
		if br.local.capture == nil {
			return
		}
		if !br.local.capture.Stopped() {
			br.local.capture.Quit()
		}
	case inputTypeYoutube:
		if !br.youtube.handler.Stopped() {
			br.youtube.handler.Quit()

		}
	}
}

// Wait waits for the bridge to finish.
func (br *Bridge) Wait() {
	<-br.done
}
