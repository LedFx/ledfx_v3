package youtube

import (
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"io"
	"ledfx/audio"
	"os"
	"sync"
	"time"
)

type Player struct {
	mu *sync.Mutex

	done    *atomic.Bool
	paused  *atomic.Bool
	unpause chan bool
	playing *atomic.Bool

	in  *FileBuffer
	out *audio.AsyncMultiWriter

	elapsed *atomic.Duration
	ticker  *time.Ticker
}

func (p *Player) Reset(input *FileBuffer) {
	p.mu.Lock()
	p.done.Store(true)

	defer func() {
		p.done.Store(false)
		p.mu.Unlock()
	}()

	p.Stop()
	p.in = input
	p.elapsed.Store(0)
}

func (p *Player) elapsedLoop(done chan struct{}) {
	if p.ticker == nil {
		p.ticker = time.NewTicker(time.Second)
	} else {
		p.ticker.Reset(time.Second)
	}

	defer p.ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-p.ticker.C:
			if !p.paused.Load() {
				p.elapsed.Add(1 * time.Second)
			}
		}
	}
}

func (p *Player) Start() error {
	p.mu.Lock()
	p.playing.Store(true)
	defer func() {
		p.playing.Store(false)
		p.mu.Unlock()
	}()

	doneCh := make(chan struct{})

	defer func() {
		doneCh <- struct{}{}
		close(doneCh)
	}()

	go p.elapsedLoop(doneCh)

	for {
		switch {
		case p.paused.Load():
			<-p.unpause
		case p.done.Load():
			return nil
		default:
			if _, err := io.CopyN(p.out, p.in, 1408); err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.ErrShortBuffer) && !errors.Is(err, os.ErrClosed) {
					return fmt.Errorf("unexpected error copying to output writer: %w", err)
				}
				return nil
			}
		}
	}
}

func (p *Player) Pause() {
	p.paused.Store(true)
}

func (p *Player) Unpause() {
	p.paused.Store(false)
	p.unpause <- true
}

func (p *Player) Stop() {
	if p.in != nil {
		p.in.Close()
	}
}

func (p *Player) IsPlaying() bool {
	return p.playing.Load()
}

func (p *Player) Close() error {
	if p.paused.Load() {
		p.unpause <- true
	}
	p.done.Store(true)
	if p.in == nil {
		return nil
	} else {
		return p.in.Close()
	}
}
