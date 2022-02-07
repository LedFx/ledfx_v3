package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	log "ledfx/logger"
	"os"
	"sync"
)

type IntWriter interface {
	Write(b Buffer) (n int, err error)
}
type IntReader interface {
	Read(p Buffer) (n int, err error)
}
type NamedMultiWriter struct {
	mu             *sync.Mutex
	writers        []io.Writer
	indexMap       map[string]int
	asyncThreshold int
	writeFn        func(p []byte) (n int, err error)
	wg             *sync.WaitGroup
}

func NewByteWriter() *NamedMultiWriter {
	nmw := &NamedMultiWriter{
		mu:             &sync.Mutex{},
		writers:        make([]io.Writer, 0),
		indexMap:       make(map[string]int),
		asyncThreshold: 2,
		wg:             &sync.WaitGroup{},
	}

	nmw.writeFn = nmw.writeSeq
	return nmw
}

func (bw *NamedMultiWriter) SetAsyncThreshold(threshold int) {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	bw.asyncThreshold = threshold

	bw.checkAsyncThreshold()
}

func (bw *NamedMultiWriter) checkAsyncThreshold() {
	// We only check if it's equal since checking for >= would spam the log.
	if len(bw.writers) == bw.asyncThreshold {
		log.Logger.WithField("category", "Audio MultiWriter").Infof("Writer threshold reached! (Writers: %d || Threshold: %d)", len(bw.writers), bw.asyncThreshold)
		log.Logger.WithField("category", "Audio MultiWriter").Infoln("Enabling asynchronous streaming to compensate for threshold delay...")
		bw.writeFn = bw.writeAsync
	} else {
		bw.writeFn = bw.writeSeq
	}
}

// AddWriter adds a writer and ties the writer index to the provided name.
func (bw *NamedMultiWriter) AddWriter(writer io.Writer, name string) error {
	if name == "" {
		return NameCannotBeOmitted
	}
	bw.mu.Lock()
	defer bw.mu.Unlock()
	bw.writers = append(bw.writers, writer)
	bw.indexMap[name] = len(bw.writers) - 1

	bw.checkAsyncThreshold()
	return nil
}

// RemoveWriter removes the writer corresponding with the provided name.
//
// Name cannot be omitted.
func (bw *NamedMultiWriter) RemoveWriter(name string) error {
	if name == "" {
		return NameCannotBeOmitted
	}
	bw.mu.Lock()
	defer bw.mu.Unlock()

	index, ok := bw.indexMap[name]
	if !ok {
		return WriterNotFound
	}

	bw.writers = append(bw.writers[:index], bw.writers[index+1:]...)
	delete(bw.indexMap, name)
	for key, val := range bw.indexMap {
		if val > index {
			bw.indexMap[key]--
		}
	}

	bw.checkAsyncThreshold()

	return nil
}

// RemoveAll removes all writers referenced by (bw *NamedMultiWriter).
func (bw *NamedMultiWriter) RemoveAll() {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	bw.writers = bw.writers[:0]
	bw.indexMap = make(map[string]int)

	bw.checkAsyncThreshold()
}

func (bw *NamedMultiWriter) Write(p []byte) (int, error) {
	return bw.writeFn(p)
}

func (bw *NamedMultiWriter) writeAsync(p []byte) (int, error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	bw.wg.Add(len(bw.writers))

	for i := range bw.writers {
		go func(i2 int) {
			defer bw.wg.Done()
			if _, err := bw.writers[i2].Write(p); err != nil {
				log.Logger.WithField("category", "Named MultiWriter").Errorf("Error writing to writer with index %d: %v", i2, err)
			}
		}(i)
	}
	bw.wg.Wait()
	return len(p), nil
}

func (bw *NamedMultiWriter) writeSeq(p []byte) (int, error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	for i := range bw.writers {
		if n, err := bw.writers[i].Write(p); err != nil {
			return n, err
		}
	}
	return len(p), nil
}

type Buffer []int16

func (b Buffer) AsFloat64() []float64 {
	out := make([]float64, len(b))
	for i, x := range b {
		out[i] = float64(x)
	}
	return out
}

func (b Buffer) AsBytes() []byte {
	byteBuf := make([]byte, len(b)*2)

	var offset int
	for i := range b {
		binary.LittleEndian.PutUint16(byteBuf[offset:], uint16(b[i]))
		offset += 2
	}
	return byteBuf
}

func (b Buffer) Sum() int64 {
	var sum int64
	for i := range b {
		sum += int64(b[i])
	}
	return sum
}

func (b Buffer) ChannelMultiplier(numChannels int) Buffer {
	b2 := Buffer(make([]int16, len(b)*numChannels))
	for i := range b {
		magnitude := b[i]
		for channel := 0; channel < numChannels; channel++ {
			b2[(i*numChannels)+channel] = magnitude
		}
	}
	return b2
}

func (b Buffer) Mono2Stereo() Buffer {
	stereo := Buffer(make([]int16, len(b)*2))
	for i := range b {
		magnitude := b[i]
		stereo[i*2] = magnitude
		stereo[(i*2)+1] = magnitude
	}
	return stereo
}

func (b Buffer) WriteTo(filename string) error {
	fi, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		return fmt.Errorf("error creating file '%s': %w", filename, err)
	}
	defer fi.Close()
	for i := range b {
		if _, err := fi.WriteString(fmt.Sprintf("%d\n", b[i])); err != nil {
			return fmt.Errorf("error writing string to file: %w", err)
		}
	}
	return nil
}

func (b Buffer) HighestValue() int16 {
	var highest int16
	for i := range b {
		if b[i] > highest {
			highest = b[i]
		}
	}
	return highest
}

const (
	dBMax = float64(96.33)
	dBMin = float64(0)

	rawMax = int16(32767)
	rawMin = int16(-32768)
)

func (b Buffer) Decibels() []float64 {
	out := make([]float64, len(b))
	for i := range b {
		switch b[i] {
		case rawMax:
			out[i] = dBMax
		case rawMin:
			out[i] = dBMin
		default:
			out[i] = (float64(b[i]) / float64(rawMax)) * dBMax
		}
	}
	return out
}

func HighestFloat(f []float64) float64 {
	var highest float64
	for i := range f {
		if f[i] > highest {
			highest = f[i]
		}
	}
	return highest
}
