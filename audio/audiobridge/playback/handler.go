package playback

import (
	"fmt"
	"github.com/dustin/go-broadcast"
	"github.com/gordonklaus/portaudio"
	"io"
	"ledfx/audio"
	"ledfx/config"
)

type Handler struct {
	*portaudio.Stream
	buffer      audio.Buffer
	byteWriters []io.Writer
	hermes      broadcast.Broadcaster
	hermesChan  chan interface{}
}

func NewHandler(audioDevice config.AudioDevice, hermes broadcast.Broadcaster, byteWriters ...io.Writer) (h *Handler, err error) {
	dev, err := audio.GetPaDeviceInfo(audioDevice)
	if err != nil {
		return nil, fmt.Errorf("error getting PortAudio device info: %w", err)
	}

	p := portaudio.StreamParameters{
		Output: portaudio.StreamDeviceParameters{
			Device:   dev,
			Channels: dev.MaxInputChannels,
		},
		SampleRate:      dev.DefaultSampleRate,
		FramesPerBuffer: int(dev.DefaultSampleRate / 60),
	}

	h = &Handler{
		buffer:      audio.Buffer{},
		byteWriters: byteWriters,
		hermes:      hermes,
		hermesChan:  make(chan interface{}),
	}

	if h.Stream, err = portaudio.OpenStream(p, h.buffer); err != nil {
		return nil, fmt.Errorf("error opening PortAudio: %w", err)
	}

	if err = h.Stream.Start(); err != nil {
		return nil, fmt.Errorf("error starting playback stream: %w", err)
	}

	go func() {
		h.hermes.Register(h.hermesChan)
		defer h.hermes.Unregister(h.hermesChan)
		for msg := range h.hermesChan {
			h.buffer = msg.(audio.Buffer)
			h.Stream.Write()
		}
	}()

	return h, nil
}

func (h *Handler) Quit() {
	h.Stream.Stop()
	h.Stream.Close()
	h.Stream = nil

	h.hermes.Unregister(h.hermesChan)
	close(h.hermesChan)
	h.buffer = nil

}
