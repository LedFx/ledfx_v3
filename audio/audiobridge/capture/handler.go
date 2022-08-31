package capture

import (
	"fmt"
	"ledfx/audio"
	log "ledfx/logger"

	"github.com/gordonklaus/portaudio"
)

type Handler struct {
	*portaudio.Stream
	byteWriter *audio.AsyncMultiWriter
	stopped    bool
}

func NewHandler(id string, byteWriter *audio.AsyncMultiWriter) (h *Handler, err error) {
	audioDevice, err := audio.GetDeviceByID(id)
	if err != nil {
		return nil, err
	}
	log.Logger.WithField("context", "Local Capture Init").Debugf("Getting info for device '%s'...", audioDevice.Name)
	dev, err := audio.GetPaDeviceInfo(audioDevice)
	if err != nil {
		return nil, fmt.Errorf("error getting PortAudio device info: %w", err)
	}

	p := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   dev,
			Channels: dev.MaxInputChannels,
		},
		SampleRate:      dev.DefaultSampleRate,
		FramesPerBuffer: int(dev.DefaultSampleRate / 60),
	}

	h = &Handler{
		byteWriter: byteWriter,
	}

	switch p.Input.Channels {
	case 1:
		log.Logger.WithField("context", "Local Capture Init").Debugf("Opening stream with Mono2Stereo callback...")
		if h.Stream, err = portaudio.OpenStream(p, h.monoCallback); err != nil {
			return nil, fmt.Errorf("error opening mono Portaudio stream: %w", err)
		}
	case 2:
		log.Logger.WithField("context", "Local Capture Init").Debugf("Opening stream with Stereo callback...")
		if h.Stream, err = portaudio.OpenStream(p, h.stereo2monoCallback); err != nil {
			return nil, fmt.Errorf("error opening stereo Portaudio stream: %w", err)
		}
	default:
		return nil, fmt.Errorf("%d channel audio is unsupported (LedFX only supports stereo/mono)", p.Input.Channels)
	}

	log.Logger.WithField("context", "Local Capture Init").Debugf("Starting stream...")
	if err = h.Stream.Start(); err != nil {
		return nil, fmt.Errorf("error starting capture stream: %w", err)
	}

	return h, nil
}

func (h *Handler) stereo2monoCallback(in audio.Buffer) {
	h.monoCallback(in.Stereo2Mono())
}

func (h *Handler) monoCallback(in audio.Buffer) {
	h.byteWriter.Write(in.AsBytes())
}

func (h *Handler) Quit() {
	h.stopped = true
	log.Logger.WithField("context", "Capture Handler").Warnf("Aborting stream...")
	h.Stream.Abort()
	log.Logger.WithField("context", "Capture Handler").Warnf("Closing stream...")
	h.Stream.Close()
}

func (h *Handler) Stopped() bool {
	return h.stopped
}
