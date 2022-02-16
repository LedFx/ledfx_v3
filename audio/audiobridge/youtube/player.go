package youtube

import (
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
	p.done = true
	defer func() {
		p.done = false
	}()
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Stop()
	p.in = input
}

func (p *Player) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	buf := make([]byte, 1408)
	for {
		switch {
		case p.paused:
			<-p.unpause
		case p.done:
			return nil
		default:
			n, err := io.ReadAtLeast(p.in, buf, 1408)
			if err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.ErrShortBuffer) && !errors.Is(err, os.ErrClosed) {
					return fmt.Errorf("unexpected error copying to output writer: %w", err)
				}
				return nil
			}

			if p.intOut != nil {
				if _, err := p.intOut.Write(audio.BytesToAudioBuffer(buf[:n][:])); err != nil {
					return fmt.Errorf("error writing to int writer: %w", err)
				}
			}

			if _, err := p.out.Write(buf[:n][:]); err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
					return fmt.Errorf("unexpected error copying to output writer: %w", err)
				}
				return nil
			}
			if n < 1408 {
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

func (p *Player) Stop() {
	if p.in != nil {
		p.in.Close()
	}
}

func (p *Player) Close() error {
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
