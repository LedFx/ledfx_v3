package audiobridge

import (
	"github.com/gordonklaus/portaudio"
	"github.com/hajimehoshi/oto"
	"io"
	"ledfx/integrations/airplay2"
	"ledfx/integrations/bluetooth"
)

// Bridge is a type for configuring the bridging of two audio devices.
type Bridge struct {
	ledFxWriter io.Writer

	SourceEndpoint *EndpointConfig `json:"source_endpoint"`
	DestEndpoint   *EndpointConfig `json:"dest_endpoint"`

	done chan bool

	airplayServer *airplay2.Server
	airplayClient *airplay2.Client

	localAudioCtx  *oto.Context
	localAudioDest *oto.Player

	localAudioSourceDone chan struct{}
	localAudioSource     *portaudio.Stream

	bluetoothClient *bluetooth.Client
	bluetoothServer *bluetooth.Server
}

// DeviceType constants
type DeviceType string

const (
	DeviceTypeAirPlay   DeviceType = "AIRPLAY"
	DeviceTypeBluetooth DeviceType = "BLUETOOTH"
)

type EndpointConfig struct {
	// Type specifies the type of source/destination device
	Type DeviceType `json:"type"`

	// IP is only applicable to AirPlay destination devices.
	// It takes priority over Name, if used in Bridge.DestEndpoint.
	IP string `json:"ip"`

	// Name is applicable to both AirPlay and Bluetooth source/dest devices.
	//
	// For destination devices, (i.e. devices that require discovery)
	// it is interpreted as a regex string.
	//
	// For source devices, (i.e. servers that are spun up by LedFX)
	// it is interpreted as a literal and no string-manipulation
	// or pattern matching will occur.
	Name string `json:"name"`

	// Mac is only applicable to Bluetooth destination devices.
	Mac string `json:"mac"`

	// Verbose, if true, prints all sorts of debug information
	// that may be valuable/insightful.
	Verbose bool `json:"verbose"`
}
