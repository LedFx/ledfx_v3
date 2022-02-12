package audiobridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"ledfx/audio/audiobridge/youtube"
	"strings"
)

type JsonCTL struct {
	w *BridgeJSONWrapper

	// YouTube stuff
	curYouTubePlaylistPlayer *youtube.PlaylistPlayer
	curYouTubePlayer         *youtube.Player
	curYouTubePlayerType     youTubePlayerType
}

func (w *BridgeJSONWrapper) CTL() *JsonCTL {
	if w.jsonCTL != nil {
		return w.jsonCTL
	} else {
		w.jsonCTL = &JsonCTL{
			w:                    w,
			curYouTubePlayerType: -1,
		}
		return w.CTL()
	}
}

type youTubePlayerType int

const (
	youTubePlayerTypeSingle youTubePlayerType = iota
	youTubePlayerTypePlaylist
)

type YouTubeAction int

const (
	YouTubeActionDownload YouTubeAction = iota
	YouTubeActionPlay
	YouTubeActionPause
	YouTubeActionResume
	YouTubeActionNext
	YouTubeActionPrevious
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

	if conf.Action != YouTubeActionDownload && j.curYouTubePlayerType == -1 {
		return fmt.Errorf("download must be called before any other YouTube control statements")
	}

	if conf.Action > 5 || conf.Action < 0 {
		return fmt.Errorf("'action' must be between 0 and 5")
	}

	switch conf.Action {
	case YouTubeActionDownload:
		if strings.Contains(strings.ToLower(conf.URL), "&list=") {
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
			return j.curYouTubePlaylistPlayer.Next()
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
			return errors.New("playlist required for 'Next' and 'Previous' action types")
		case youTubePlayerTypePlaylist:
			return j.curYouTubePlaylistPlayer.Next()
		}
	case YouTubeActionPrevious:
		switch j.curYouTubePlayerType {
		case youTubePlayerTypeSingle:
			return errors.New("playlist required for 'Next' and 'Previous' action types")
		case youTubePlayerTypePlaylist:
			return j.curYouTubePlaylistPlayer.Previous()
		}
	}
	return nil
}
