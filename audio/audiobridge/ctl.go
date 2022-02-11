package audiobridge

import "ledfx/audio/audiobridge/youtube"

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
	return ytc.handler.handler.PlayPlaylist(playlistURL)
}
func (ytc *YoutubeController) Play(videoURL string) (*youtube.Player, error) {
	return ytc.handler.handler.Play(videoURL)
}

// --- END YOUTUBE CTL ---

func (c *Controller) Local() *LocalController {
	return &LocalController{
		handler: c.br.local,
	}
}

func (c *Controller) AirPlay() *AirPlayController {
	return &AirPlayController{
		handler: c.br.airplay,
	}
}

type YoutubeController struct {
	handler *YoutubeHandler
}
type LocalController struct {
	handler *LocalHandler
}
type AirPlayController struct {
	handler *AirPlayHandler
}
