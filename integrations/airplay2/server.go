package airplay2

import (
	"ledfx/audio"
	"ledfx/handlers/raop"
	log "ledfx/logger"
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

func NewServer(conf Config, intWriter audio.IntWriter, byteWriter *audio.AsyncMultiWriter) (s *Server) {
	pl := newPlayer(intWriter, byteWriter)

	if conf.Port == 0 {
		conf.Port = 7000
	}

	if conf.AdvertisementName == "" {
		conf.AdvertisementName = "LedFX"
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

func (s *Server) AddClient(client *Client) error {
	return s.player.AddClient(client)
}

func (s *Server) Start() error {
	errCh := make(chan error)
	go func() {
		defer func() {
			s.done <- struct{}{}
		}()
		err := s.svc.Start(s.conf.VerboseLogging, true)
		errCh <- err
		if err != nil {
			log.Logger.WithField("category", "AirPlay Server").Errorf("Error starting AirPlay server: %v", err)
			return
		}
		s.svc.Wait()
	}()
	return <-errCh
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
