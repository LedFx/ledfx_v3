package capture

import (
	"fmt"

	"github.com/LedFx/ledfx/pkg/audio"
	log "github.com/LedFx/ledfx/pkg/logger"

	"github.com/LedFx/portaudio"
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
			Channels: 1,
			Latency:  0,
		},
		Output:          portaudio.StreamDeviceParameters{},
		SampleRate:      dev.DefaultSampleRate,
		FramesPerBuffer: int(audio.BufferSize),
	}

	h = &Handler{
		byteWriter: byteWriter,
	}

	log.Logger.WithField("context", "Local Capture Init").Debugf("Opening stream...")
	if h.Stream, err = portaudio.OpenStream(p, h.monoCallback); err != nil {
		return nil, fmt.Errorf("error opening stream: %w", err)
	}

	log.Logger.WithField("context", "Local Capture Init").Debugf("Starting stream...")
	if err = h.Stream.Start(); err != nil {
		return nil, fmt.Errorf("error starting stream: %w", err)
	}

	return h, nil
}

func (h *Handler) monoCallback(in audio.Buffer) {
	h.byteWriter.Write(in.AsBytes())
}

func (h *Handler) Quit() {
	h.stopped = true
	log.Logger.WithField("context", "Capture Handler").Debug("Aborting stream...")
	h.Stream.Abort()
	log.Logger.WithField("context", "Capture Handler").Debug("Closing stream...")
	h.Stream.Close()
	log.Logger.WithField("context", "Capture Handler").Info("Closed stream")
}

func (h *Handler) Stopped() bool {
	return h.stopped
}
