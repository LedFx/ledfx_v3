package youtube

import (
	"errors"
	"fmt"
	yt "github.com/kkdai/youtube/v2"
	log "ledfx/logger"
)

type PlaylistPlayer struct {
	h        *Handler
	trackNum int
	tracks   []string
}

func (pp *PlaylistPlayer) Next() error {
	pp.inc()
	p, err := pp.h.Play(pp.tracks[pp.trackNum])
	if err != nil {
		if errors.Is(err, yt.ErrNotPlayableInEmbed) {
			log.Logger.WithField("category", "YT Playlist Player").Warnf("Could not play track %d: %v", pp.trackNum, err)
			return pp.Next()
		}
		return fmt.Errorf("error playing track %d: %w", pp.trackNum, err)
	}
	go func() {
		if err := p.Start(); err != nil {
			log.Logger.WithField("category", "YT Playlist Player").Errorf("Error starting playback: %v", err)
		}
	}()
	return nil
}

func (pp *PlaylistPlayer) Previous() error {
	pp.dec()
	p, err := pp.h.Play(pp.tracks[pp.trackNum])
	if err != nil {
		if errors.Is(err, yt.ErrNotPlayableInEmbed) {
			log.Logger.WithField("category", "YT Playlist Player").Warnf("Could not play track %d: %v", pp.trackNum, err)
			return pp.Previous()
		}
		return fmt.Errorf("error playing track %d: %w", pp.trackNum, err)
	}
	go func() {
		if err := p.Start(); err != nil {
			log.Logger.WithField("category", "YT Playlist Player").Warnf("Error starting playback: %v", err)
		}
	}()
	return nil
}

func (pp *PlaylistPlayer) inc() {
	if pp.trackNum >= len(pp.tracks)-1 {
		pp.trackNum = 0
	} else {
		pp.trackNum++
	}
}

func (pp *PlaylistPlayer) dec() {
	if pp.trackNum <= 0 {
		pp.trackNum = len(pp.tracks) - 1
	} else {
		pp.trackNum--
	}
}

func (pp *PlaylistPlayer) NumTracks() int {
	return len(pp.tracks) - 1
}
func (pp *PlaylistPlayer) PlayTrackNum(num int) error {
	if num >= len(pp.tracks)-1 || num < 0 {
		return fmt.Errorf("track number must be between 0 and %d", len(pp.tracks)-1)
	}

	pp.trackNum = num
	p, err := pp.h.Play(pp.tracks[pp.trackNum])
	if err != nil {
		if errors.Is(err, yt.ErrNotPlayableInEmbed) {
			log.Logger.WithField("category", "YT Playlist Player").Warnf("Could not play track %d: %v", pp.trackNum, err)
			return pp.Next()
		}
		return fmt.Errorf("error playing track %d: %w", pp.trackNum, err)
	}
	go func() {
		if err := p.Start(); err != nil {
			log.Logger.WithField("category", "YT Playlist Player").Warnf("Error starting playback: %v", err)
		}
	}()
	return nil
}

func (pp *PlaylistPlayer) Stop() {
	pp.trackNum = -1
	pp.tracks = pp.tracks[:0]
}
