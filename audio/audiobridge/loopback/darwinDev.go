//go:build darwin || windows

package loopback

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
)

type Dev struct {
	e        *echoExperimental
	loopback *Loopback
	done     chan bool
}

func (d *Dev) Stop() {
	//portaudio.Terminate()
	d.e.Close()
	//d.e.Stop()
	d.done <- true
}

func (d *Dev) Wait() {
	<-d.done
}

func NewDev(l *Loopback) (d *Dev, err error) {
	if err = portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("error initializing portaudio: %w", err)
	}

	d = &Dev{
		loopback: l,
		done:     make(chan bool),
	}

	if d.e, err = echoExp(l); err != nil {
		return nil, fmt.Errorf("error initializing new echo device: %w", err)
	}
	
	return d, nil
}

/* TODO Definitely make it possible to serve the Portaudio stream over an HTTP endpoint. */
