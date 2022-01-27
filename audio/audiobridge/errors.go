package audiobridge

import "errors"

var (
	ErrCannotBridgeBT2AP = errors.New("invalid configuration: cannot bridge audio from Bluetooth to AirPlay")
)
