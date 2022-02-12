package audiobridge

import (
	"ledfx/audio"
)

// Bridge can wire up an audio source to multiple destinations
// seamlessly and with minimal delay.
type Bridge struct {
	inputType inputType

	bufferCallback func(buf audio.Buffer)
	byteWriter     *audio.AsyncMultiWriter
	intWriter      audio.IntWriter

	airplay *AirPlayHandler
	local   *LocalHandler
	youtube *YoutubeHandler

	ctl *Controller

	done chan bool

	jsonWrapper *BridgeJSONWrapper
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

// CallbackWrapper wraps a buffer Callback into a struct
type CallbackWrapper struct {
	Callback func(buf audio.Buffer)
}

// BridgeJSONWrapper wraps a bridge with a JSON interpreter
type BridgeJSONWrapper struct {
	br      *Bridge
	jsonCTL *JsonCTL
}
