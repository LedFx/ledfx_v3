package audiobridge

import (
	"encoding/json"
	"ledfx/config"
)

// AirPlayInputJSON configures an AirPlay input (server)
type AirPlayInputJSON struct {
	Name    string `json:"name"`
	Port    int    `json:"port"`
	Verbose bool   `json:"verbose"`
}

func (a AirPlayInputJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&a)
}

// AirPlayOutputJSON configures an AirPlay output
type AirPlayOutputJSON struct {
	SearchKey  string            `json:"search_key"`
	SearchType AirPlaySearchType `json:"search_type"`
	Verbose    bool              `json:"verbose"`
}

func (a AirPlayOutputJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&a)
}

// LocalInputJSON configures a local input (capture)
type LocalInputJSON struct {
	AudioDevice *config.AudioDevice `json:"audio_device,omitempty"`
	Verbose     bool                `json:"verbose"`
}

func (l LocalInputJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&l)
}

// LocalOutputJSON configures a local output (playback)
type LocalOutputJSON struct {
	Verbose bool `json:"verbose"`
}

func (l LocalOutputJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&l)
}
