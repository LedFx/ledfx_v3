package audiobridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"ledfx/audio/audiobridge/youtube"
	log "ledfx/logger"
)

type YouTubeAction string

const (
	// YouTubeActionDownload
	//
	// If supplied with a playlist URL, the handler downloads
	// all corresponding audio track(s) from each playlist entry.
	//
	// If supplied with a video URL, the handler downloads the
	// audio track from the provided video.
	YouTubeActionDownload YouTubeAction = "download"

	// YouTubeActionPlay plays all tracks until the end of the queue is reached.
	// Upon completion, YouTubeActionResume must be called to restart from the beginning.
	YouTubeActionPlay = "play"

	// YouTubeActionPause pauses playback
	YouTubeActionPause = "pause"

	// YouTubeActionResume resumes/unpauses playback
	YouTubeActionResume = "resume"

	// YouTubeActionStop stops the handler, closes all playback, and clears the queue.
	// This should NOT be used for pausing.
	YouTubeActionStop = "stop"

	// YouTubeActionNext skips the current track.
	YouTubeActionNext = "next"

	// YouTubeActionPrevious rewinds to the previous track.
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
func (j *JsonCTL) YouTubeSet(jsonData []byte) (respBytes []byte, err error) {
	conf := YouTubeCTLJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	switch {
	case j.w.br.youtube == nil:
		fallthrough
	case j.w.br.youtube.handler == nil:
		return nil, errors.New("YouTube handler is nil")
	default:
		j.curYouTubePlayer = j.w.br.youtube.handler.Player()
	}

	switch conf.Action {
	case YouTubeActionDownload:
		log.Logger.WithField("category", "YouTube JSONCTL").Infof("Downloading audio track(s) from URL '%s'", conf.URL)
		return nil, j.curYouTubePlayer.Download(conf.URL)
	case YouTubeActionPlay:
		log.Logger.WithField("category", "YouTube JSONCTL").Infof("Starting YouTube playback...")
		if err := j.curYouTubePlayer.Play(); err != nil {
			return nil, err
		}
		return json.Marshal(j.curYouTubePlayer.NowPlaying())
	case YouTubeActionStop:
		log.Logger.WithField("category", "YouTube JSONCTL").Infof("Stopping YouTube player...")
		return nil, j.curYouTubePlayer.Close()
	case YouTubeActionPause:
		log.Logger.WithField("category", "YouTube JSONCTL").Infof("Pausing YouTube playback...")
		j.curYouTubePlayer.Pause()
	case YouTubeActionResume:
		log.Logger.WithField("category", "YouTube JSONCTL").Infof("Resuming YouTube playback...")
		j.curYouTubePlayer.Unpause()
		return json.Marshal(j.curYouTubePlayer.NowPlaying())
	case YouTubeActionNext:
		log.Logger.WithField("category", "YouTube JSONCTL").Infof("Skipping to next YouTube track...")
		j.curYouTubePlayer.Next()
		return json.Marshal(j.curYouTubePlayer.NowPlaying())
	case YouTubeActionPrevious:
		log.Logger.WithField("category", "YouTube JSONCTL").Infof("Rewinding to previous YouTube track...")
		j.curYouTubePlayer.Previous()
		return json.Marshal(j.curYouTubePlayer.NowPlaying())
	default:
		return nil, fmt.Errorf("unknown action '%s'", conf.Action)
	}
	return nil, nil
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
