package capture

import (
	"fmt"
	"ledfx/audio"
	"ledfx/config"
	log "ledfx/logger"

	"github.com/gordonklaus/portaudio"
)

type Handler struct {
	*portaudio.Stream
	byteWriter *audio.AsyncMultiWriter
	verbose    bool
	stopped    bool
}

func NewHandler(audioDevice config.AudioDevice, byteWriter *audio.AsyncMultiWriter, verbose bool) (h *Handler, err error) {
	if verbose {
		log.Logger.WithField("context", "Local Capture Init").Infof("Getting info for device '%s'...", audioDevice.Name)
	}
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
		verbose:    verbose,
	}

	switch p.Input.Channels {
	case 1:
		if verbose {
			log.Logger.WithField("context", "Local Capture Init").Infof("Opening stream with Mono2Stereo callback...")
		}
		if h.Stream, err = portaudio.OpenStream(p, h.mono2StereoCallback); err != nil {
			return nil, fmt.Errorf("error opening mono Portaudio stream: %w", err)
		}
	case 2:
		if verbose {
			log.Logger.WithField("context", "Local Capture Init").Infof("Opening stream with Stereo callback...")
		}
		if h.Stream, err = portaudio.OpenStream(p, h.stereoCallback); err != nil {
			return nil, fmt.Errorf("error opening stereo Portaudio stream: %w", err)
		}
	default:
		return nil, fmt.Errorf("%d channel audio is unsupported (LedFX only supports stereo/mono)", p.Input.Channels)
	}

	if verbose {
		log.Logger.WithField("context", "Local Capture Init").Infof("Starting stream...")
	}
	if err = h.Stream.Start(); err != nil {
		return nil, fmt.Errorf("error starting capture stream: %w", err)
	}

	return h, nil
}

func (h *Handler) stereoCallback(in audio.Buffer) {
	h.byteWriter.Write(in.AsBytes())
}

func (h *Handler) mono2StereoCallback(in audio.Buffer) {
	h.stereoCallback(in.ChannelMultiplier(2))
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
