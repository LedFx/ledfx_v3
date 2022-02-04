package hermeshandler

import (
	"fmt"
	"io"
	"ledfx/audio"
	log "ledfx/logger"
	"unsafe"
)

type Interpreter struct {
	ch chan interface{}
}

func NewInterpreter(hermesChan chan interface{}) *Interpreter {
	return &Interpreter{hermesChan}
}

func (i *Interpreter) NextAsBytes() ([]byte, error) {
	switch val, ok := <-i.ch; p := val.(type) {
	case []byte:
		return p, nil
	case audio.Buffer:
		return p.AsBytes(), nil
	default:
		if !ok {
			return nil, io.EOF
		}
		log.Logger.WithField("category", "Hermes Handler").Errorf("Unimplemented broadcast type '%T'!", val)
		return nil, fmt.Errorf("unimplemented broadcast type '%T'", val)
	}
}

func (i *Interpreter) NextAsBuffer() (out audio.Buffer, err error) {
	switch val, ok := <-i.ch; p := val.(type) {
	case []byte:
		out = audio.Buffer{}
		var offset int
		for i := 0; i < len(p); i += 2 {
			out = append(out, readInt16Unsafe(p[i:i+2]))
			offset++
		}
		return out, nil
	case audio.Buffer:
		return p, nil
	default:
		if !ok {
			return nil, io.EOF
		}
		log.Logger.WithField("category", "Hermes Handler").Errorf("Unimplemented broadcast type '%T'!", val)
		return nil, fmt.Errorf("unimplemented broadcast type '%T'", val)
	}
}

func readInt16Unsafe(b []byte) int16 {
	return *(*int16)(unsafe.Pointer(&b[0]))
}
