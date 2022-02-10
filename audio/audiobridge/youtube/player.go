package youtube

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"ledfx/audio"
	"os"
	"sync"
)

type Player struct {
	mu *sync.Mutex

	done    bool
	paused  bool
	unpause chan bool

	in     *os.File
	out    *audio.AsyncMultiWriter
	intOut audio.IntWriter
}

func (p *Player) Reset(input *os.File) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.in != nil {
		p.in.Close()
		p.in = nil
	}
	p.in = input
}

func (p *Player) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	buf := bytes.NewBuffer(make([]byte, 1408))
	for {
		switch {
		case p.paused:
			<-p.unpause
		case p.done:
			return nil
		default:
			if _, err := buf.ReadFrom(io.LimitReader(p.in, 1408)); err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
					return fmt.Errorf("unexpected error copying to output writer: %w", err)
				}
				return nil
			}

			if _, err := p.intOut.Write(audio.BytesToAudioBuffer(buf.Bytes())); err != nil {
				return fmt.Errorf("error writing to int writer: %w", err)
			}

			if _, err := buf.WriteTo(p.out); err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
					return fmt.Errorf("unexpected error copying to output writer: %w", err)
				}
				return nil
			}
		}
	}
}

func (p *Player) Pause() {
	p.paused = true
}

func (p *Player) Unpause() {
	p.paused = false
	p.unpause <- true
}

func (p *Player) Stop() error {
	if p.paused {
		p.unpause <- true
	}
	p.done = true
	if p.in == nil {
		return nil
	} else {
		return p.in.Close()
	}
}
