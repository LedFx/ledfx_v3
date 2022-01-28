package audiobridge

import (
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/hajimehoshi/oto"
	aubio "github.com/simonassank/aubio-go"
	"io"
	"ledfx/audio"
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
	callbackHandler      *CallbackHandler

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

type CallbackHandler struct {
	buf        *aubio.SimpleBuffer
	pvoc       *aubio.PhaseVoc
	melbank    *aubio.FilterBank
	onset      *aubio.Onset
	frameCount uint
	fpb        uint
	writeTo    io.Writer
}

func NewCallbackHandler() (c *CallbackHandler, err error) {
	c = &CallbackHandler{
		frameCount: 60,
		fpb:        44100 / 60,
	}

	if c.pvoc, err = aubio.NewPhaseVoc(1024, c.fpb); err != nil {
		return nil, err
	}

	c.melbank = aubio.NewFilterBank(40, 1024)
	c.melbank.SetMelCoeffsSlaney(uint(44100))

	if c.onset, err = aubio.NewOnset(aubio.HFC, 1024, c.fpb, 44100); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *CallbackHandler) audioCallback(in audio.Buffer) {
	c.frameCount += 1
	c.buf = aubio.NewSimpleBufferData(uint(len(in)), audio.BufferToF64(&in))
	defer c.buf.Free()
	c.pvoc.Do(c.buf)
	c.melbank.Do(c.pvoc.Grain())
	c.onset.Do(c.buf)

	if c.onset.Buffer().Slice()[0] != 0 {
		fmt.Println("nice clap!")
	}
}

func (c *CallbackHandler) setWriteTo(writeTo io.Writer) {
	c.writeTo = writeTo
}
