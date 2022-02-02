package capture

import (
	"encoding/binary"
	"fmt"
	"github.com/dustin/go-broadcast"
	"github.com/gordonklaus/portaudio"
	"io"
	"ledfx/audio"
	"ledfx/config"
	log "ledfx/logger"
)

type Handler struct {
	*portaudio.Stream
	byteWriters []io.Writer
	hermes      broadcast.Broadcaster
}

func (h *Handler) AddByteWriters(byteWriters ...io.Writer) {
	h.byteWriters = append(h.byteWriters, byteWriters...)
}

func NewHandler(audioDevice config.AudioDevice, hermes broadcast.Broadcaster, byteWriters ...io.Writer) (h *Handler, err error) {
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
		byteWriters: byteWriters,
		hermes:      hermes,
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
	h.hermes.Submit(in)
	byteBuf := make([]byte, len(in)*2)

	var offset int
	for i := range in {
		binary.LittleEndian.PutUint16(byteBuf[offset:], uint16(in[i]))
		offset += 2
	}

	for i := range h.byteWriters {
		_, _ = h.byteWriters[i].Write(byteBuf)
	}
}

func (h *Handler) monoToStereoCallback(in audio.Buffer) {
	h.stereoCallback(in.Mono2Stereo())
}

func (h *Handler) Quit() {
	log.Logger.WithField("category", "Capture Handler").Warnf("Aborting stream...")
	h.Stream.Abort()
	log.Logger.WithField("category", "Capture Handler").Warnf("Closing stream...")
	h.Stream.Close()
}
