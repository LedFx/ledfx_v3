package airplay2

import (
	"fmt"
	"github.com/carterpeel/bobcaygeon/raop"
	"io"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func init() {
	// Init rand seed for the server port
	rand.Seed(time.Now().UnixNano())
}

type Server struct {
	mu     sync.Mutex
	player *audioPlayer
	conf   *Config
	svc    *raop.AirplayServer

	done chan struct{}

	runLock *atomic.Value
}

func NewServer(conf Config) (s *Server) {
	s = &Server{
		mu:      sync.Mutex{},
		conf:    &conf,
		player:  newPlayer(), // Port range: 1024 through 65530
		runLock: &atomic.Value{},
		done:    make(chan struct{}),
	}

	s.runLock.Store(false)

	return s
}

func (s *Server) AddOutput(output io.Writer) {
	s.player.AddWriter(output)
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.runLock.Load().(bool) {
		return fmt.Errorf("cannot acquire AirPlay server lock")
	}

	s.runLock.Store(true)
	s.svc = raop.NewAirplayServer(rand.Intn(65530-1024+1)+1024, s.conf.AdvertisementName, s.player) //nolint:gosec
	go func() {
		defer func() {
			s.done <- struct{}{}
		}()
		s.svc.Start(s.conf.VerboseLogging, true)
	}()

	return nil
}

func (s *Server) Wait() {
	<-s.done
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.runLock.Load().(bool) {
		return fmt.Errorf("cannot stop inactive AirPlay server")
	}

	s.runLock.Store(false)
	s.svc.Stop()

	return nil
}
