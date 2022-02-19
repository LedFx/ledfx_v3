package playback

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"io"
	"ledfx/audio"
	log "ledfx/logger"
	"ledfx/util"
)

type Player struct {
	identifier string
	stream     *portaudio.Stream
	outDev     *portaudio.DeviceInfo
	buf        audio.Buffer
	verbose    bool
	done       bool
}

func NewHandler(verbose bool) (h *Player, err error) {
	h = &Player{
		identifier: util.RandString(8),
		verbose:    verbose,
		buf:        make([]int16, 1408/2),
	}

	if h.outDev, err = portaudio.DefaultOutputDevice(); err != nil {
		return nil, fmt.Errorf("error getting default output device: %w", err)
	}

	if verbose {
		log.Logger.WithField("category", "Local Playback Init").Infof("Default output device: %s", h.outDev.Name)
		log.Logger.WithField("category", "Local Capture Init").Infof("Opening stream... (%dCH/16-bit @%vhz)", h.outDev.MaxOutputChannels, h.outDev.DefaultSampleRate)
	}

	// Ensure format compatibility with the data sent over Player.Write()
	h.outDev.MaxOutputChannels = 2
	h.outDev.DefaultSampleRate = 44100

	if h.stream, err = portaudio.OpenStream(portaudio.LowLatencyParameters(nil, h.outDev), &h.buf); err != nil {
		return nil, fmt.Errorf("error opening PortAudio stream: %w", err)
	}
	if verbose {
		log.Logger.WithField("category", "Local Capture Init").Infof("Starting stream...")
	}
	if err = h.stream.Start(); err != nil {
		return nil, fmt.Errorf("error starting stream: %w", err)
	}
	return h, nil
}

func (wh *Player) Write(p []byte) (n int, err error) {
	if wh.done {
		return 0, io.EOF
	}
	copy(wh.buf, audio.BytesToAudioBuffer(p))
	_ = wh.stream.Write()
	return len(p), nil
}

func (wh *Player) Quit() {
	if wh.stream != nil {
		wh.stream.Abort()
		wh.stream = nil
		wh.done = true
	}
}

func (wh *Player) Identifier() string {
	return wh.identifier
}
