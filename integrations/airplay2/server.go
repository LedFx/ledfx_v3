package airplay2

import (
	"github.com/carterpeel/bobcaygeon/raop"
	"io"
	"math/rand"
	"sync"
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
}

func NewServer(conf Config) (s *Server) {
	pl := newPlayer()

	if conf.Port == 0 {
		conf.Port = 7000
	}

	s = &Server{
		mu:     sync.Mutex{},
		conf:   &conf,
		player: pl, // Port range: 1024 through 65530
		done:   make(chan struct{}),
		svc:    raop.NewAirplayServer(conf.Port, conf.AdvertisementName, pl),
	}

	return s
}

func (s *Server) AddOutput(output io.Writer) {
	s.player.AddWriter(output)
}

func (s *Server) SetClient(client *Client) {
	s.player.SetClient(client)
}

func (s *Server) Start() error {
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

func (s *Server) Stop() {
	if s.svc != nil {
		s.svc.Stop()
	}
	if s.player != nil {
		s.player.Close()
	}
}
