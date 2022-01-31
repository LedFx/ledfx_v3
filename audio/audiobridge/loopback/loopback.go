package loopback

import (
	"fmt"
	"io"
)

type Loopback struct {
	numOutputs int
	outputs    []io.Writer
	dev        *Dev
}

func New(outputs ...io.Writer) (l *Loopback, err error) {
	l = &Loopback{
		outputs: outputs,
	}
	if l.dev, err = NewDev(l); err != nil {
		return nil, fmt.Errorf("error initializing new loopback device: %w", err)
	}
	return l, nil
}

func (l *Loopback) AddOutput(output io.Writer) {
	l.outputs = append(l.outputs, output)
	l.numOutputs++
}

func (l *Loopback) Wait() {
	l.dev.Wait()
}
func (l *Loopback) Done() {
	l.dev.Stop()
}
