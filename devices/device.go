package device

import "ledfx/color"

// All devices take pixels and send them somewhere
type PixelPusher interface {
	Send(p *color.Pixels)
	Connect()
	Disconnect()
}

type Device struct {
	ID          string
	Pusher      PixelPusher
	Dimensioner PixelDimensioner
	// BaseConfig  BaseDeviceConfig
}

type State int

const (
	Offline State = iota
	Connecting
	Disconnecting
	Connected
)

func (s State) String() string {
	switch s {
	case Offline:
		return "offline"
	case Connecting:
		return "connecting"
	case Disconnecting:
		return "disconnecting"
	case Connected:
		return "connected"
	}
	return "unknown"
}
