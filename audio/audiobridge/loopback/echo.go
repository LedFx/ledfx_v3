package loopback

import (
	"encoding/binary"
	"fmt"
	"github.com/gordonklaus/portaudio"
	conv "github.com/oov/audio/converter"
	"github.com/zaf/resample"
	"github.com/zikichombo/sio"
	"github.com/zikichombo/sound"
	"github.com/zikichombo/sound/sample"
	"io"
	log "ledfx/logger"
	"math"
)

type echoExperimental struct {
	sound.Source
	rsmp *resample.Resampler
}

type echo struct {
	*portaudio.Stream
	buffer []float32
	i      int
}

func (e *echo) processAudio(in, out []float32) {
	for i := range in {
		out[i] = .7 * e.buffer[e.i]
		e.buffer[e.i] = in[i]
		e.i = (e.i + 1) % len(e.buffer)
	}
}

func newEcho() (*echo, error) {
	h, err := portaudio.DefaultHostApi()
	if err != nil {
		return nil, fmt.Errorf("error getting default Portaudio host API: %w", err)
	}

	log.Logger.Infof("Input device: %s\n", h.Devices[0].Name)
	log.Logger.Infof("Output device: %s\n", h.Devices[2].Name)

	p := portaudio.LowLatencyParameters(h.Devices[0], h.Devices[2])
	p.Input.Channels = 2
	p.Output.Channels = 2
	e := &echo{buffer: make([]float32, int(p.SampleRate/8))}
	if e.Stream, err = portaudio.OpenStream(p, e.processAudio); err != nil {
		return nil, fmt.Errorf("error opening Portaudio stream: %w", err)
	}
	return e, nil
}

func echoExp(l *Loopback) (e *echoExperimental, err error) {
	e = new(echoExperimental)

	if e.Source, err = sio.CaptureWith(sound.StereoCd(), sample.Codecs[10], 64); err != nil {
		return nil, fmt.Errorf("error opening audio capture source: %w", err)
	}

	floatBuf := make([]float64, 64)

	outWriter := io.MultiWriter(l.outputs...)

	go func() {
		for {
			if _, err := e.Source.Receive(floatBuf); err != nil {
				log.Logger.Errorf("error recieving audio: %v", err)
			}

			byteBuf := make([]byte, 64*8)
			conv.Float64.FromFloat64(floatBuf, byteBuf)

			if _, err := outWriter.Write(byteBuf); err != nil {
				log.Logger.Errorf("error streaming to output writer: %v", err)
			}
		}
	}()
	return e, nil
}

func FloatsToInts(floats []float64) []int16 {
	ints := make([]int16, len(floats))
	for i := range ints {
		ints[i] = int16(floats[i] * 32767.0)
	}
	return ints
}

func FloatsToBytes(floats []float64) []byte {
	buf := make([][]byte, len(floats))
	for i := range floats {
		buf[i] = make([]byte, 8)
		binary.BigEndian.PutUint64(buf[i][:], math.Float64bits(floats[i]))
	}
	out := make([]byte, 0)
	for i := range buf {
		out = append(out, buf[i]...)
	}
	return out
}

func float64ToByte(f float64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], math.Float64bits(f))
	return buf[:]
}
