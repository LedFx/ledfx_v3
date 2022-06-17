package youtube

/*
import (
	"encoding/json"
	"errors"
	"fmt"
	yt "github.com/kkdai/youtube/v2"
	log "ledfx/logger"
	"time"
)

type SongDuration time.Duration

func (d SongDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

type TrackInfo struct {
	Artist        string       `json:"artist,omitempty"`
	Title         string       `json:"title,omitempty"`
	Duration      SongDuration `json:"duration,omitempty"`
	SampleRate    int64        `json:"samplerate,omitempty"`
	FileSize      int64        `json:"filesize,omitempty"`
	URL           string       `json:"url,omitempty"`
	AudioChannels int          `json:"audio_channels,omitempty"`

	invalid bool
}

type PlaylistPlayer struct {
	h        *Handler
	trackNum int
	tracks   []TrackInfo
}

func (pp *PlaylistPlayer) Pause() {
	pp.h.p.Pause()
}
func (pp *PlaylistPlayer) Unpause() {
	pp.h.p.Unpause()
}

func (pp *PlaylistPlayer) Next(waitDone bool) error {
	pp.inc()
	switch {
	case len(pp.tracks) <= pp.trackNum:
		return fmt.Errorf("no tracks found with index %d", pp.trackNum)
	case pp.tracks[pp.trackNum].invalid:
		return pp.Next(waitDone)
	}

	p, err := pp.h.Play(pp.tracks[pp.trackNum].URL)
	if err != nil {
		if errors.Is(err, yt.ErrNotPlayableInEmbed) {
			log.Logger.WithField("context", "YT Playlist Player").Warnf("Could not play track %d: %v", pp.trackNum, err)
			return pp.Next(waitDone)
		}
		return fmt.Errorf("error playing track %d: %w", pp.trackNum, err)
	}

	if waitDone {
		return p.Start()
	} else {
		go func() {
			if err := p.Start(); err != nil {
				log.Logger.WithField("context", "YT Playlist Player").Errorf("Error starting playback: %v", err)
			}
		}()
	}

	return nil
}

func (pp *PlaylistPlayer) Previous(waitDone bool) error {
	pp.dec()

	switch {
	case len(pp.tracks) <= pp.trackNum:
		return fmt.Errorf("no tracks found with index %d", pp.trackNum)
	case pp.tracks[pp.trackNum].invalid:
		return pp.Next(waitDone)
	}

	p, err := pp.h.Play(pp.tracks[pp.trackNum].URL)
	if err != nil {
		if errors.Is(err, yt.ErrNotPlayableInEmbed) {
			log.Logger.WithField("context", "YT Playlist Player").Warnf("Could not play track %d: %v", pp.trackNum, err)
			return pp.Previous(waitDone)
		}
		return fmt.Errorf("error playing track %d: %w", pp.trackNum, err)
	}

	if waitDone {
		return p.Start()
	} else {
		go func() {
			if err := p.Start(); err != nil {
				log.Logger.WithField("context", "YT Playlist Player").Warnf("Error starting playback: %v", err)
			}
		}()
	}
	return nil
}

func (pp *PlaylistPlayer) inc() {
	if pp.trackNum >= len(pp.tracks) {
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
func (pp *PlaylistPlayer) PlayTrackNum(num int, waitDone bool) error {
	if num >= len(pp.tracks)-1 || num < 0 {
		return fmt.Errorf("track number must be between 0 and %d", len(pp.tracks)-1)
	}

	pp.trackNum = num
	p, err := pp.h.Play(pp.tracks[pp.trackNum].URL)
	if err != nil {
		if errors.Is(err, yt.ErrNotPlayableInEmbed) {
			log.Logger.WithField("context", "YT Playlist Player").Warnf("Could not play track %d: %v", pp.trackNum, err)
			return pp.Next(waitDone)
		}
		return fmt.Errorf("error playing track %d: %w", pp.trackNum, err)
	}
	go func() {
		if err := p.Start(); err != nil {
			log.Logger.WithField("context", "YT Playlist Player").Warnf("Error starting playback: %v", err)
		}
	}()
	return nil
}

func (pp *PlaylistPlayer) Stop() {
	pp.trackNum = -1
	pp.tracks = pp.tracks[:0]
	pp.StopCurrentTrack()
}

func (pp *PlaylistPlayer) StopCurrentTrack() {
	if pp.h != nil {
		if pp.h.p != nil {
			pp.h.p.Stop()
		}
	}
}
*/
