package playback

import (
	"fmt"
	"io"
	"ledfx/audio"
	log "ledfx/logger"
	"ledfx/util"

	"github.com/gordonklaus/portaudio"
)

type WindowsHandler struct {
	identifier string
	stream     *portaudio.Stream
	outDev     *portaudio.DeviceInfo
	buf        audio.Buffer
	done       bool
}

func (wh *WindowsHandler) Device() string {
	return wh.outDev.Name
}

func (wh *WindowsHandler) SampleRate() int {
	return int(wh.outDev.DefaultSampleRate)
}

func (wh *WindowsHandler) NumChannels() int8 {
	return int8(wh.outDev.MaxOutputChannels)
}

func (wh *WindowsHandler) CurrentBufferSize() int {
	return len(wh.buf)
}

func NewHandler() (h *WindowsHandler, err error) {
	h = &WindowsHandler{
		identifier: util.RandString(8),
		buf:        make([]int16, 1408/2),
	}

	// todo choose an output device using identifier like input device
	if h.outDev, err = portaudio.DefaultOutputDevice(); err != nil {
		return nil, fmt.Errorf("error getting default output device: %w", err)
	}

	log.Logger.WithField("context", "Local Playback Init").Debugf("Default output device: %s", h.outDev.Name)
	log.Logger.WithField("context", "Local Capture Init").Debugf("Opening stream... (%dCH/16-bit @%vhz)", h.outDev.MaxOutputChannels, h.outDev.DefaultSampleRate)

	// Ensure format compatibility with the data sent over Player.Write()
	h.outDev.DefaultSampleRate = 44100
	h.outDev.MaxOutputChannels = 2

	if h.stream, err = portaudio.OpenDefaultStream(
		0,
		h.outDev.MaxOutputChannels,
		h.outDev.DefaultSampleRate,
		int(h.outDev.DefaultSampleRate/60),
		h.buf,
	); err != nil {
		return nil, fmt.Errorf("error opening PortAudio stream: %w", err)
	}
	log.Logger.WithField("context", "Local Capture Init").Debug("Starting stream...")
	if err = h.stream.Start(); err != nil {
		return nil, fmt.Errorf("error starting stream: %w", err)
	}
	return h, nil
}

func (wh *WindowsHandler) Write(p []byte) (n int, err error) {
	if wh.done {
		return 0, io.EOF
	}
	copy(wh.buf, audio.BytesToAudioBuffer(p))
	_ = wh.stream.Write()
	return len(p), nil
}

func (wh *WindowsHandler) Quit() {
	if wh.stream != nil {
		wh.stream.Abort()
		wh.stream = nil
		wh.done = true
	}
}

func (wh *WindowsHandler) Identifier() string {
	return wh.identifier
}
