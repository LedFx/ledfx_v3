package audiobridge

import "errors"

var (
	ErrCannotBridgeBT2AP                 = errors.New("invalid configuration: cannot bridge audio from Bluetooth to AirPlay")
	ErrAirplayRequiredForArtworkGradient = errors.New("A source device type of 'DeviceTypeAirPlay' is required to retrieve artwork gradients")
)
