package audiobridge

import (
	"ledfx/audio"
)

// Bridge is a type for configuring the bridging of two audio devices.
type Bridge struct {
	inputType inputType

	bufferCallback func(buf audio.Buffer)
	byteWriter     *audio.ByteWriter
	intWriter      audio.IntWriter

	airplay *AirPlayHandler
	local   *LocalHandler

	done chan bool
}

type inputType int8

const (
	inputTypeAirPlayServer inputType = iota
	inputTypeLocal
)

type callbackWrapper struct {
	callback func(buf audio.Buffer)
}
