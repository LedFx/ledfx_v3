package audiobridge

import (
	"go.uber.org/atomic"
	"ledfx/audio/audiobridge/youtube"
)

type JsonCTL struct {
	w *BridgeJSONWrapper

	// YouTube stuff
	curYouTubePlaylistPlayer *youtube.PlaylistPlayer
	curYouTubePlayer         *youtube.Player
	curYouTubePlayerType     youTubePlayerType
	keepPlaying              *atomic.Bool

	// AirPlay stuff
}

func (w *BridgeJSONWrapper) CTL() *JsonCTL {
	if w.jsonCTL != nil {
		return w.jsonCTL
	} else {
		w.jsonCTL = &JsonCTL{
			w:                    w,
			curYouTubePlayerType: -1,
			keepPlaying:          atomic.NewBool(false),
		}
		return w.CTL()
	}
}
