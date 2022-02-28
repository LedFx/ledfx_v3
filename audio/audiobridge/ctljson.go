package audiobridge

import (
	"ledfx/audio/audiobridge/youtube"
)

type JsonCTL struct {
	w *BridgeJSONWrapper

	// YouTubeSet stuff
	curYouTubePlayer *youtube.Player

	// AirPlay stuff
}

func (w *BridgeJSONWrapper) CTL() *JsonCTL {
	if w.jsonCTL != nil {
		return w.jsonCTL
	} else {
		w.jsonCTL = &JsonCTL{
			w: w,
		}
		return w.CTL()
	}
}
