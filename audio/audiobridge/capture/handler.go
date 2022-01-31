package capture

import (
	"encoding/binary"
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/oov/audio/converter"
	"io"
	"ledfx/audio"
	"ledfx/config"
	log "ledfx/logger"
)

type Handler struct {
	*portaudio.Stream
	buffer      audio.Buffer
	byteWriters []io.Writer
	done        chan bool
	floatChan   chan []float64
}

func (h *Handler) AddByteWriters(byteWriters ...io.Writer) {
	h.byteWriters = append(h.byteWriters, byteWriters...)
}

func NewHandler(audioDevice config.AudioDevice, floatChan chan []float64, byteWriters ...io.Writer) (*Handler, error) {
	h, err := portaudio.DefaultHostApi()
	if err != nil {
		return nil, fmt.Errorf("error getting default Portaudio host API: %w", err)
	}

	log.Logger.Infof("Input device: %s\n", h.DefaultInputDevice.Name)

	dev, err := audio.GetPaDeviceInfo(audioDevice)
	if err != nil {
		return nil, fmt.Errorf("error getting Pulseaudio device info: %w", err)
	}

	p := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   dev,
			Channels: dev.MaxInputChannels,
		},
		SampleRate:      dev.DefaultSampleRate,
		FramesPerBuffer: int(dev.DefaultSampleRate / 60),
	}

	e := &Handler{
		buffer:      audio.Buffer{},
		done:        make(chan bool),
		byteWriters: byteWriters,
		floatChan:   floatChan,
	}

	if e.Stream, err = portaudio.OpenStream(p, e.audioSampleCallback); err != nil {
		return nil, fmt.Errorf("error opening Portaudio stream: %w", err)
	}

	e.Stream.Start()

	return e, nil
}

func (h *Handler) audioSampleCallback(in audio.Buffer) {
	byteBuf := make([]byte, len(in)*2)
	floatBuf := make([]float64, len(in))

	for i := range in {
		floatBuf[i] = converter.Int16ToFloat64(in[i])
	}

	h.floatChan <- floatBuf

	var offset int
	for i := range in {
		binary.LittleEndian.PutUint16(byteBuf[offset:], uint16(in[i]))
		offset += 2
	}

	for i := range h.byteWriters {
		_, _ = h.byteWriters[i].Write(byteBuf)
	}
}

func (h *Handler) Wait() {
	<-h.done
	h.Stream.Stop()
	h.Stream.Close()
	portaudio.Terminate()
}

func (h *Handler) Quit() {
	h.done <- true
}

/*func echoExp(l *Loopback) (e *echoExperimental, err error) {
	e = new(echoExperimental)

	if e.Source, err = sio.CaptureWith(sound.StereoCd(), sample.Codecs[2], 64); err != nil {
		return nil, fmt.Errorf("error opening audio capture source: %w", err)
	}

	floatBuf := make([]float64, 64)
	outWriter := io.MultiWriter(l.outputs...)

	go func() {
		for {
			if _, err := e.Source.Receive(floatBuf); err != nil {
				log.Logger.Errorf("error recieving audio: %v", err)
			}

			intBuf := make([]int16, 64)

			for i := range floatBuf {
				intBuf[i] = int16(floatBuf[i] * 32767)

			}

			if _, err := outWriter.Write(byteBuf); err != nil {
				log.Logger.Errorf("error streaming to output writer: %v", err)
			}
		}
	}()
	return e, nil
}*/
