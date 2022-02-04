package capture

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"ledfx/audio"
	"ledfx/config"
	log "ledfx/logger"
)

type Handler struct {
	*portaudio.Stream
	intWriter  audio.IntWriter
	byteWriter *audio.ByteWriter
}

func NewHandler(audioDevice config.AudioDevice, intWriter audio.IntWriter, byteWriter *audio.ByteWriter) (h *Handler, err error) {
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
		intWriter:  intWriter,
		byteWriter: byteWriter,
	}

	switch p.Input.Channels {
	case 1:
		if h.Stream, err = portaudio.OpenStream(p, h.monoToStereoCallback); err != nil {
			return nil, fmt.Errorf("error opening mono Portaudio stream: %w", err)
		}
	case 2:
		if h.Stream, err = portaudio.OpenStream(p, h.stereoCallback); err != nil {
			return nil, fmt.Errorf("error opening stereo Portaudio stream: %w", err)
		}
	default:
		return nil, fmt.Errorf("%d channel audio is unsupported (LedFX only supports stereo/mono)", p.Input.Channels)
	}

	if err = h.Stream.Start(); err != nil {
		return nil, fmt.Errorf("error starting capture stream: %w", err)
	}

	return h, nil
}

func (h *Handler) stereoCallback(in audio.Buffer) {
	h.byteWriter.Write(in.AsBytes())
	h.intWriter.Write(in)
}

func (h *Handler) monoToStereoCallback(in audio.Buffer) {
	h.stereoCallback(in.ChannelMultiplier(2))
}

func (h *Handler) Quit() {
	log.Logger.WithField("category", "Capture Handler").Warnf("Aborting stream...")
	h.Stream.Abort()
	log.Logger.WithField("category", "Capture Handler").Warnf("Closing stream...")
	h.Stream.Close()
}
