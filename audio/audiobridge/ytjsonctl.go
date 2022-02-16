package audiobridge

import (
	"encoding/json"
	"errors"
	"fmt"
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
	YouTubeActionDownload YouTubeAction = "download"
	YouTubeActionPlay                   = "play"
	YouTubeActionPause                  = "pause"
	YouTubeActionResume                 = "resume"

	// YouTubeActionNext only applies to playlists
	YouTubeActionNext = "next"
	// YouTubeActionPrevious only applies to playlists
	YouTubeActionPrevious = "previous"
	// YouTubeActionPlayAll only applies to playlists
	YouTubeActionPlayAll = "playall"
)

type YouTubeCTLJSON struct {
	Action YouTubeAction `json:"action"`
	URL    string        `json:"url"`
}

func (ytctl YouTubeCTLJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&ytctl)
}

// YouTube takes a marshalled YouTubeCTLJSON
func (j *JsonCTL) YouTube(jsonData []byte) (err error) {
	conf := YouTubeCTLJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	switch {
	case conf.Action == YouTubeActionDownload && conf.URL == "":
		return errors.New("action 'Download' requires the 'url' field to be populated")
	case conf.Action != YouTubeActionDownload && j.curYouTubePlayerType == -1:
		return errors.New("download must be called before any other YouTube control statements")
	}

	switch conf.Action {
	case YouTubeActionDownload:
		j.keepPlaying.Store(false)
		if strings.Contains(strings.ToLower(conf.URL), "list=") {
			j.curYouTubePlayerType = youTubePlayerTypePlaylist
			if j.curYouTubePlaylistPlayer, err = j.w.br.Controller().YouTube().PlayPlaylist(conf.URL); err != nil {
				return fmt.Errorf("error playing playlist: %w", err)
			}
		} else {
			j.curYouTubePlayerType = youTubePlayerTypeSingle
			if j.curYouTubePlayer, err = j.w.br.Controller().YouTube().Play(conf.URL); err != nil {
				return fmt.Errorf("error playing url: %w", err)
			}
		}
	case YouTubeActionPlay:
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			return j.curYouTubePlayer.Start()
		case youTubePlayerTypePlaylist:
			j.keepPlaying.Store(false)
			return j.curYouTubePlaylistPlayer.Next(false)
		}
	case YouTubeActionPause:
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			j.curYouTubePlayer.Pause()
		case youTubePlayerTypePlaylist:
			j.curYouTubePlaylistPlayer.Pause()
		}
	case YouTubeActionResume:
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			j.curYouTubePlayer.Unpause()
		case youTubePlayerTypePlaylist:
			j.curYouTubePlaylistPlayer.Unpause()
		}
	case YouTubeActionNext:
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			return errors.New("playlist required for 'Next', 'Previous' and 'PlayAll' action types")
		case youTubePlayerTypePlaylist:
			return j.curYouTubePlaylistPlayer.Next(false)
		}
	case YouTubeActionPrevious:
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			return errors.New("playlist required for 'Next', 'Previous' and 'PlayAll' action types")
		case youTubePlayerTypePlaylist:
			return j.curYouTubePlaylistPlayer.Previous(false)
		}
	case YouTubeActionPlayAll:
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			return errors.New("playlist required for 'Next', 'Previous' and 'PlayAll' action types")
		case youTubePlayerTypePlaylist:
			j.keepPlaying.Store(true)
			go func() {
				for j.keepPlaying.Load() {
					if err := j.curYouTubePlaylistPlayer.Next(true); err != nil {
						log.Logger.WithField("category", "YouTube JSON Handler").Errorf("Error playing next playlist track: %v", err)
					}
				}
			}()
			return nil
		}
	default:
		return fmt.Errorf("unknown action '%s'", conf.Action)
	}
	return nil
}
