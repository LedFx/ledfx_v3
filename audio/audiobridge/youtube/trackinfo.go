package youtube

import (
	"encoding/json"
	"github.com/kkdai/youtube/v2"
	"time"
)

type TrackInfo struct {
	Artist        string       `json:"artist,omitempty"`
	Title         string       `json:"title,omitempty"`
	Duration      SongDuration `json:"duration,omitempty"`
	SampleRate    int64        `json:"samplerate,omitempty"`
	FileSize      int64        `json:"filesize,omitempty"`
	URL           string       `json:"url,omitempty"`
	AudioChannels int          `json:"audio_channels,omitempty"`

	// Invalid states whether an error occurred during the download process.
	Invalid bool `json:"invalid,omitempty"`

	video *youtube.Video
}
type SongDuration time.Duration

func (d SongDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}
