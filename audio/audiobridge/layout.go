package audiobridge

import (
	"github.com/dustin/go-broadcast"
	"ledfx/audio"
)

// Bridge is a type for configuring the bridging of two audio devices.
type Bridge struct {
	inputType inputType

	bufferCallback func(buf audio.Buffer)
	hermes         broadcast.Broadcaster

	airplay *AirPlayHandler
	local   *LocalHandler
}

type inputType int8

const (
	inputTypeAirPlayServer inputType = iota
	inputTypeLocal
)
