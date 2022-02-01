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

	if h.Stream, err = portaudio.OpenStream(p, h.captureSampleCallback); err != nil {
		return nil, fmt.Errorf("error opening Portaudio stream: %w", err)
	}

	if err = h.Stream.Start(); err != nil {
		return nil, fmt.Errorf("error starting capture stream: %w", err)
	}

	return h, nil
}

func (h *Handler) captureSampleCallback(in audio.Buffer) {
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

func (h *Handler) Quit() {
	log.Logger.WithField("category", "Capture Handler").Warnf("Aborting stream...")
	h.Stream.Abort()
	log.Logger.WithField("category", "Capture Handler").Warnf("Closing stream...")
	h.Stream.Close()
}
