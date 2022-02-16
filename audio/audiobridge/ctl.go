package audiobridge

import (
	"fmt"
	"ledfx/audio/audiobridge/youtube"
	"ledfx/integrations/airplay2"
)

func (br *Bridge) Controller() *Controller {
	return br.ctl
}

func (br *Bridge) newController() *Controller {
	return &Controller{
		br: br,
	}
}

type Controller struct {
	br *Bridge
}

// --- BEGIN YOUTUBE CTL ---

// YouTube returns a YoutubeController
func (c *Controller) YouTube() *YoutubeController {
	return &YoutubeController{
		handler: c.br.youtube,
	}
}
func (ytc *YoutubeController) PlayPlaylist(playlistURL string) (*youtube.PlaylistPlayer, error) {
	if ytc.handler != nil {
		if ytc.handler.handler != nil {
			return ytc.handler.handler.PlayPlaylist(playlistURL)
		}
	}
	return nil, fmt.Errorf("YouTube playback is not active")
}

func (ytc *YoutubeController) Play(videoURL string) (*youtube.Player, error) {
	if ytc.handler != nil {
		if ytc.handler.handler != nil {
			return ytc.handler.handler.Play(videoURL)
		}
	}
	return nil, fmt.Errorf("YouTube playback is not active")
}

// --- END YOUTUBE CTL ---

// --- BEGIN LOCAL CTL ---

// Local returns a *LocalController
func (c *Controller) Local() *LocalController {
	return &LocalController{
		handler: c.br.local,
	}
}

func (lc *LocalController) Stop() error {
	if lc.handler != nil {
		lc.handler.Stop()
		return nil
	}
	return fmt.Errorf("local handler is not active")
}
func (lc *LocalController) SetVerbose(enabled bool) error {
	if lc.handler != nil {
		lc.handler.verbose = enabled
		return nil
	}
	return fmt.Errorf("local handler is not active")
}
func (lc *LocalController) QuitPlayback() error {
	if lc.handler != nil {
		if lc.handler.playback != nil {
			lc.handler.playback.Quit()
			return nil
		}
	}
	return fmt.Errorf("local playback is not active")
}
func (lc *LocalController) QuitCapture() error {
	if lc.handler != nil {
		if lc.handler.capture != nil {
			lc.handler.capture.Quit()
			return nil
		}
	}
	return fmt.Errorf("local capture is not active")
}
func (lc *LocalController) PlaybackIdentifier() (string, error) {
	if lc.handler != nil {
		if lc.handler.playback != nil {
			return lc.handler.playback.Identifier(), nil
		}
	}
	return "", fmt.Errorf("local playback is not active")
}

// --- END LOCAL CTL ---

// --- BEGIN AIRPLAY CTL ---

// AirPlay returns an *AirPlayController
func (c *Controller) AirPlay() *AirPlayController {
	return &AirPlayController{
		handler: c.br.airplay,
	}
}

func (apc *AirPlayController) StopServer() error {
	if apc.handler != nil {
		if apc.handler.server != nil {
			apc.handler.server.Stop()
			return nil
		}
	}
	return fmt.Errorf("server is not active")
}
func (apc *AirPlayController) Clients() []*airplay2.Client {
	if apc.handler != nil {
		return apc.handler.clients
	}
	return nil
}

// --- END AIRPLAY CTL ---

type YoutubeController struct {
	handler *YoutubeHandler
}
type LocalController struct {
	handler *LocalHandler
}
type AirPlayController struct {
	handler *AirPlayHandler
}
