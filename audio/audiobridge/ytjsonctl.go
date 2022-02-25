package audiobridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/atomic"
	"ledfx/audio/audiobridge/youtube"
	log "ledfx/logger"
	"strings"
)

type youTubePlayerType int

const (
	youTubePlayerTypeSingle youTubePlayerType = iota
	youTubePlayerTypePlaylist
)

type YouTubeAction string

const (
	// YouTubeActionDownload stores the requested URL and prepares the handler to play it
	YouTubeActionDownload YouTubeAction = "download"
	// YouTubeActionPlay plays the requested URL or playlist
	YouTubeActionPlay = "play"
	// YouTubeActionPause pauses playback
	YouTubeActionPause = "pause"
	// YouTubeActionResume resumes/unpauses playback
	YouTubeActionResume = "resume"
	// YouTubeActionStop stops the handler, closes all playback, and clears the queue.
	// This should NOT be used for pausing.
	YouTubeActionStop = "stop"

	// YouTubeActionNext only applies to playlists
	YouTubeActionNext = "next"
	// YouTubeActionPrevious only applies to playlists
	YouTubeActionPrevious = "previous"
)

type YouTubeCTLJSON struct {
	Action YouTubeAction `json:"action"`
	URL    string        `json:"url"`
}

func (ytctl YouTubeCTLJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&ytctl)
}

// YouTubeSet takes a marshalled YouTubeCTLJSON
func (j *JsonCTL) YouTubeSet(jsonData []byte) (err error) {
	conf := YouTubeCTLJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	switch {
	case conf.Action == YouTubeActionDownload && conf.URL == "":
		return errors.New("action 'Download' requires the 'url' field to be populated")
	case conf.Action != YouTubeActionDownload && j.curYouTubePlayerType == -1:
		return errors.New("download must be called before any other YouTubeSet control statements")
	}

	switch conf.Action {
	case YouTubeActionDownload:
		j.keepPlaying.Store(false)
		if strings.Contains(strings.ToLower(conf.URL), "list=") {
			j.curYouTubePlayerType = youTubePlayerTypePlaylist
			log.Logger.WithField("category", "YouTube JSONCTL").Infof("Downloading all audio tracks from playlist URL '%s'", conf.URL)
			if j.curYouTubePlaylistPlayer, err = j.w.br.Controller().YouTube().PlayPlaylist(conf.URL); err != nil {
				return fmt.Errorf("error downloading playlist: %w", err)
			}
		} else {
			j.curYouTubePlayerType = youTubePlayerTypeSingle
			log.Logger.WithField("category", "YouTube JSONCTL").Infof("Downloading audio from video URL '%s'", conf.URL)
			if j.curYouTubePlayer, err = j.w.br.Controller().YouTube().Play(conf.URL); err != nil {
				return fmt.Errorf("error downloading url: %w", err)
			}
		}
	case YouTubeActionPlay:
		log.Logger.WithField("category", "YouTube JSONCTL").Infof("Starting YouTube playback...")
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			if j.curYouTubePlayer.IsPlaying() {
				return errors.New("already playing")
			}
			return j.curYouTubePlayer.Start()
		case youTubePlayerTypePlaylist:
			if j.keepPlaying.Load() {
				return errors.New("already playing")
			}
			j.keepPlayingFn = func(pp *youtube.PlaylistPlayer) error {
				return pp.Next(true)
			}
			j.keepPlaying.Store(false)
			j.curYouTubePlaylistPlayer.StopCurrentTrack()
			j.keepPlaying.Store(true)
			go j.autoPlayback(j.curYouTubePlaylistPlayer, j.keepPlaying)
			return nil
		}
	case YouTubeActionStop:
		log.Logger.WithField("category", "YouTube JSONCTL").Infof("Stopping YouTube player...")
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			j.curYouTubePlayer.Stop()
			return nil
		case youTubePlayerTypePlaylist:
			j.curYouTubePlaylistPlayer.Stop()
		}
	case YouTubeActionPause:
		log.Logger.WithField("category", "YouTube JSONCTL").Infof("Pausing YouTube playback...")
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			j.curYouTubePlayer.Pause()
		case youTubePlayerTypePlaylist:
			j.curYouTubePlaylistPlayer.Pause()
		}
	case YouTubeActionResume:
		log.Logger.WithField("category", "YouTube JSONCTL").Infof("Resuming YouTube playback...")
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			j.curYouTubePlayer.Unpause()
		case youTubePlayerTypePlaylist:
			j.curYouTubePlaylistPlayer.Unpause()
		}
	case YouTubeActionNext:
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			return errors.New("playlist required for 'next' and 'previous' action types")
		case youTubePlayerTypePlaylist:
			log.Logger.WithField("category", "YouTube JSONCTL").Infof("Skipping current YouTube playlist track...")
			j.keepPlayingFn = func(pp *youtube.PlaylistPlayer) error {
				return pp.Next(true)
			}
			j.curYouTubePlaylistPlayer.StopCurrentTrack()
			return nil
		}
	case YouTubeActionPrevious:
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			return errors.New("playlist required for 'next' and 'previous' action types")
		case youTubePlayerTypePlaylist:
			log.Logger.WithField("category", "YouTube JSONCTL").Infof("Rewinding to previous YouTube playlist track...")
			j.keepPlayingFn = func(pp *youtube.PlaylistPlayer) error {
				return pp.Previous(true)
			}
			j.curYouTubePlaylistPlayer.StopCurrentTrack()
			return nil
		}
	default:
		return fmt.Errorf("unknown action '%s'", conf.Action)
	}
	return nil
}

type YouTubeInfo struct {
	IsPlaying       bool                      `json:"is_playing"`
	PercentComplete youtube.CompletionPercent `json:"percent_complete"`
	Paused          bool                      `json:"paused"`
	TrackIndex      int                       `json:"track_index"`
	NowPlaying      youtube.TrackInfo         `json:"now_playing"`
	Queued          []youtube.TrackInfo       `json:"queued"`
}

func (ytinfo YouTubeInfo) AsJSON() ([]byte, error) {
	return json.Marshal(&ytinfo)
}

func (j *JsonCTL) YouTubeGetInfo() (resultJson []byte, err error) {
	info := YouTubeInfo{}

	if info.IsPlaying, err = j.w.br.Controller().YouTube().IsPlaying(); err != nil {
		return nil, fmt.Errorf("error getting 'IsPlaying()': %w", err)
	}
	if info.PercentComplete, err = j.w.br.Controller().YouTube().SongCompletionPercent(); err != nil {
		return nil, fmt.Errorf("error getting 'SongCompletionPercent()': %w", err)
	}
	if info.Paused, err = j.w.br.Controller().YouTube().IsPaused(); err != nil {
		return nil, fmt.Errorf("error getting 'IsPaused()': %w", err)
	}
	if info.TrackIndex, err = j.w.br.Controller().YouTube().TrackIndex(); err != nil {
		return nil, fmt.Errorf("error getting 'TrackIndex()': %w", err)
	}
	if info.NowPlaying, err = j.w.br.Controller().YouTube().NowPlaying(); err != nil {
		return nil, fmt.Errorf("error getting 'NowPlaying()': %w", err)
	}
	if info.Queued, err = j.w.br.Controller().YouTube().QueuedTracks(); err != nil {
		return nil, fmt.Errorf("error getting 'QueuedTracks()': %w", err)
	}

	return info.AsJSON()
}

func (j *JsonCTL) autoPlayback(pp *youtube.PlaylistPlayer, keepPlaying *atomic.Bool) {
	for keepPlaying.Load() {
		if err := j.keepPlayingFn(pp); err != nil {
			log.Logger.WithField("category", "YouTubeSet JSON Handler").Errorf("Error playing playlist track: %v", err)
		}
	}
}
