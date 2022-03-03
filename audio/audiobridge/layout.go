package audiobridge

import (
	"fmt"
	"ledfx/audio"
)

// Bridge can wire up an audio source to multiple destinations
// seamlessly and with minimal delay.
type Bridge struct {
	inputType inputType

	bufferCallback func(buf audio.Buffer)
	byteWriter     *audio.AsyncMultiWriter

	airplay *AirPlayHandler
	local   *LocalHandler
	youtube *YoutubeHandler

	ctl *Controller

	done chan bool

	jsonWrapper *BridgeJSONWrapper

	info *Info

	outputs []*OutputInfo
}

// inputType indicates the audio source a bridge will use
type inputType int8

const (
	// A Bridge with an inputType as inputTypeAirPlayServer will run
	// an AirPlay server for clients to connect and send audio to.
	inputTypeAirPlayServer inputType = iota

	// A Bridge with an inputType as inputTypeLocal will capture
	// system audio from a chosen local device.
	inputTypeLocal

	// A Bridge with an inputType as inputTypeYoutube will stream audio
	// from provided videos to all outputs.
	inputTypeYoutube
)

func (i inputType) String() string {
	switch i {
	case -1:
		return "UNDEFINED"
	case inputTypeAirPlayServer:
		return "AIRPLAY"
	case inputTypeLocal:
		return "CAPTURE"
	case inputTypeYoutube:
		return "YOUTUBE"
	default:
		return fmt.Sprintf("UNKNOWN:%d", i)
	}
}

// CallbackWrapper wraps a buffer Callback into a struct
type CallbackWrapper struct {
	Callback func(buf audio.Buffer)
}

// BridgeJSONWrapper wraps a bridge with a JSON interpreter
type BridgeJSONWrapper struct {
	br      *Bridge
	jsonCTL *JsonCTL
}

type OutputType string

const (
	outputTypeAirPlay   OutputType = "airplay"
	outputTypeLocal     OutputType = "local"
	outputTypeGeneric   OutputType = "generic"
	outputTypeBluetooth OutputType = "bluetooth"
)

type OutputInfo struct {
	Type OutputType  `json:"type"`
	Info interface{} `json:"info"`
}
type AirPlayOutputInfo struct {
	IP          string `json:"ip"`
	Hostname    string `json:"hostname"`
	AdvertName  string `json:"advertisement_name"`
	Type        string `json:"airplay_type"`
	DeviceModel string `json:"device_model,omitempty"`
	Port        int    `json:"port"`
	SampleRate  int    `json:"sample_rate"`
}
type LocalOutputInfo struct {
	Device     string `json:"device"`
	Identifier string `json:"identifier"`
	SampleRate int    `json:"sample_rate"`
	Channels   int8   `json:"channels"`
}
type GenericOutputInfo struct {
	Identifier string `json:"identifier"`
}
type BluetoothOutputInfo struct {
	// TODO
}
