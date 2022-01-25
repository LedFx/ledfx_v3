package airplay2

import (
	"fmt"
	"github.com/carterpeel/bobcaygeon/raop"
	"github.com/carterpeel/bobcaygeon/rtsp"
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
	reqChan chan *rtsp.Request
}

func NewServer(conf Config) (s *Server) {
	pl := newPlayer()
	s = &Server{
		mu:      sync.Mutex{},
		conf:    &conf,
		player:  pl, // Port range: 1024 through 65530
		runLock: &atomic.Value{},
		done:    make(chan struct{}),
		reqChan: make(chan *rtsp.Request),
		svc:     raop.NewAirplayServer(conf.Port, conf.AdvertisementName, pl),
	}

	if s.conf.Port == 0 {
		s.conf.Port = 35293
	}

	s.runLock.Store(false)

	return s
}

func (s *Server) AddOutput(output io.Writer) {
	s.player.AddWriter(output)
}

func (s *Server) SetClient(client *Client) {
	s.player.SetClient(client)
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.runLock.Load().(bool) {
		return fmt.Errorf("cannot acquire AirPlay server lock")
	}

	s.runLock.Store(true)
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
