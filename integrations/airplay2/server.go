package airplay2

import (
	"io"
	"ledfx/color"
	"ledfx/handlers/raop"
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

func (s *Server) AddClient(client *Client) {
	s.player.AddClient(client)
}

func (s *Server) Start() error {
	errCh := make(chan error)
	go func() {
		errCh <- s.svc.Start(s.conf.VerboseLogging, true)
		s.svc.Wait()
		defer func() {
			s.done <- struct{}{}
		}()
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

func (s *Server) GetAlbumGradient(resolution int) (*color.Gradient, error) {
	return s.player.GetGradientFromArtwork(resolution)
}

func (s *Server) AnimateArtwork(width, height, frames int) ([]byte, error) {
	return color.AnimateAlbumArt(s.player.artwork, width, height, frames)
}
