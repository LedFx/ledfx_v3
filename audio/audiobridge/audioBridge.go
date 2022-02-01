package audiobridge

import (
	"fmt"
	"github.com/dustin/go-broadcast"
	"github.com/gordonklaus/portaudio"
	"ledfx/audio"
	log "ledfx/logger"
)

// NewBridge initializes a new bridge between a source and destination audio device.
func NewBridge(bufferCallback func(buf audio.Buffer)) (br *Bridge, err error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("error initializing PortAudio: %w", err)
	}
	br = &Bridge{
		bufferCallback: bufferCallback,
		hermes:         broadcast.NewBroadcaster(60),
		inputType:      inputType(-1), // -1 signifies undefined
	}
	return br, nil
}

// Stop stops the bridge. Any further references to 'br *Bridge'
// may cause a runtime panic.
func (br *Bridge) Stop() {
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
