package audiobridge

import (
	"io"
)

// Bridge is a type for configuring the bridging of two audio devices.
type Bridge struct {
	inputType inputType

	ledFxWriter io.Writer
	done        chan bool

	airplay *AirPlayHandler
	local   *LocalHandler
}

type inputType int8

const (
	inputTypeAirPlayServer inputType = iota
	inputTypeLocal
)
