package audiobridge

import (
	"github.com/hajimehoshi/oto"
	"ledfx/integrations/airplay2"
)

// Bridge is a type for configuring the bridging of two audio devices.
type Bridge struct {
	SourceEndpoint *EndpointConfig `json:"source_endpoint"`
	DestEndpoint   *EndpointConfig `json:"dest_endpoint"`

	switchChan chan struct{}
	done       chan bool

	airplayServer *airplay2.Server
	airplayClient *airplay2.Client

	localAudioCtx  *oto.Context
	localAudioDest *oto.Player
}

// DeviceType constants
type DeviceType string

const (
	DeviceTypeAirPlay   DeviceType = "AIRPLAY"
	DeviceTypeBluetooth DeviceType = "BLUETOOTH"
	DeviceTypeLocal     DeviceType = "LOCAL"
)

type EndpointConfig struct {
	// Type specifies the type of source/destination device
	Type DeviceType `json:"type"`

	// IP is only applicable to AirPlay destination devices.
	// It takes priority over Name, if used in Bridge.DestEndpoint.
	IP string `json:"ip"`

	// Name has two functions:
	//
	// 1). MUST be populated in Bridge.SourceEndpoint (used as the name of the AirPlay server advertisement)
	//
	// 2). CAN be populated in Bridge.DestEndpoint, but will be ignored if IP is populated.
	Name string `json:"name"`

	// Mac is only applicable to Bluetooth destination devices.
	Mac string `json:"mac"`
}
