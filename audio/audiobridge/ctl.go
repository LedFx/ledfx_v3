package audiobridge

import (
	"errors"
	"fmt"
	"ledfx/audio/audiobridge/youtube"
	"ledfx/integrations/airplay2"
	"time"
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

func (c *Controller) InputType() string {
	return c.br.inputType.String()
}
func (c *Controller) Outputs() []OutputInfo {
	outputs := make([]OutputInfo, len(c.br.outputs))
	for i := range c.br.outputs {
		outputs[i] = *c.br.outputs[i]
	}
	return outputs
}

// --- BEGIN YOUTUBE CTL ---

// YouTube returns a YoutubeController
func (c *Controller) YouTube() *YoutubeController {
	return &YoutubeController{
		handler: c.br.youtube,
	}
}

func (ytc *YoutubeController) CheckErr() error {
	if ytc.handler != nil {
		if ytc.handler.handler != nil {
			return nil
		}
	}
	return fmt.Errorf("YouTube handler is not active")
}

func (ytc *YoutubeController) NowPlaying() (info youtube.TrackInfo, err error) {
	if ytc.handler != nil {
		if ytc.handler.handler != nil {
			return ytc.handler.handler.Player().NowPlaying(), nil
		}
	}
	return info, fmt.Errorf("YouTube handler is not active")
}
func (ytc *YoutubeController) QueuedTracks() ([]youtube.TrackInfo, error) {
	if ytc.handler != nil {
		if ytc.handler.handler != nil {
			return ytc.handler.handler.Player().QueuedTracks(), nil
		}
	}
	return nil, fmt.Errorf("YouTube handler is not active")
}

func (ytc *YoutubeController) TimeElapsed() (time.Duration, error) {
	if ytc.handler != nil {
		if ytc.handler.handler != nil {
			return ytc.handler.handler.Player().TimeElapsed(), nil
		}
	}
	return -1, errors.New("YouTube handler is not active")
}

func (ytc *YoutubeController) IsPaused() (bool, error) {
	if ytc.handler != nil {
		if ytc.handler.handler != nil {
			return ytc.handler.handler.Player().IsPaused(), nil
		}
	}
	return false, fmt.Errorf("YouTube handler is not active")
}
func (ytc *YoutubeController) TrackIndex() (int, error) {
	if ytc.handler != nil {
		if ytc.handler.handler != nil {
			return ytc.handler.handler.Player().TrackIndex(), nil
		}
	}
	return -1, fmt.Errorf("YouTube handler is not active")
}
func (ytc *YoutubeController) IsPlaying() (bool, error) {
	if ytc.handler != nil {
		if ytc.handler.handler != nil {
			return ytc.handler.handler.Player().IsPlaying(), nil
		}
	}
	return false, fmt.Errorf("YouTube handler is not active")
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
func (lc *LocalController) QuitPlayback() error {
	if lc.handler != nil {
		lc.handler.playback.Quit()
		return nil
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
		return lc.handler.playback.Identifier(), nil
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
func (apc *AirPlayController) Server() *airplay2.Server {
	if apc.handler != nil {
		return apc.handler.server
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
