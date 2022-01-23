package audiobridge

import (
	"fmt"
	"github.com/hajimehoshi/oto"
	"ledfx/integrations/airplay2"
)

// NewBridge initializes a new bridge between a source and destination audio device.
func NewBridge(srcConfig EndpointConfig, dstConfig EndpointConfig) (br *Bridge, err error) {
	br = &Bridge{
		SourceEndpoint: &srcConfig,
		DestEndpoint:   &dstConfig,
		switchChan:     make(chan struct{}),
		done:           make(chan bool),
	}

	if err := br.initDevices(); err != nil {
		return nil, fmt.Errorf("error initializing devices: %w", err)
	}
	return br, nil
}

func (br *Bridge) initDevices() error {
	dstType := br.DestEndpoint.Type
	srcType := br.SourceEndpoint.Type

	switch dstType {
	case DeviceTypeAirPlay:
		if err := br.initAirPlayDest(); err != nil {
			return fmt.Errorf("error initializing AirPlay destination client: %w", err)
		}
	case DeviceTypeLocal:
		if err := br.initLocalDest(); err != nil {
			return fmt.Errorf("error initializing local audio destination: %w", err)
		}
	case DeviceTypeBluetooth:
		// TODO bluetooth destination (i.e connect as client)
	}

	switch srcType {
	case DeviceTypeAirPlay:
		if err := br.initAirPlaySource(); err != nil {
			return fmt.Errorf("error initializing AirPlay server source: %w", err)
		}
	case DeviceTypeLocal:
		return fmt.Errorf("DeviceTypeLocal is not a valid audio source")
	case DeviceTypeBluetooth:
		// TODO bluetooth source (i.e advertise as server)
	}

	switch {
	case srcType == DeviceTypeAirPlay && dstType == DeviceTypeAirPlay:
		br.airplayServer.SetClient(br.airplayClient)
		if err := br.airplayServer.Start(); err != nil {
			return fmt.Errorf("error starting AirPlay server: %w", err)
		}
	case srcType == DeviceTypeAirPlay && dstType == DeviceTypeBluetooth:
		// TODO br.airplayServer.AddOutput(bluetoothWriter)
	case srcType == DeviceTypeAirPlay && dstType == DeviceTypeLocal:
		br.airplayServer.AddOutput(br.localAudioDest)
	case srcType == DeviceTypeBluetooth && dstType == DeviceTypeAirPlay:
		// TODO io.Copy(br.airplayClient, bluetoothReader)
	case srcType == DeviceTypeBluetooth && dstType == DeviceTypeBluetooth:
		// TODO io.Copy(bluetoothWriter, bluetoothReader)
	case srcType == DeviceTypeBluetooth && dstType == DeviceTypeLocal:
		// TODO io.Copy(br.localAudioDest)
	default:
		return fmt.Errorf("this error should never be returned, so if it was, we've got an issue")
	}

	return nil
}

func (br *Bridge) initAirPlaySource() error {
	br.airplayServer = airplay2.NewServer(airplay2.Config{
		AdvertisementName: "LedFX",
		Port:              7000,
		VerboseLogging:    false,
	})
	return nil
}

func (br *Bridge) initAirPlayDest() (err error) {
	if br.airplayClient, err = airplay2.NewClient(airplay2.ClientDiscoveryParameters{
		DeviceName: br.DestEndpoint.Name,
		DeviceIP:   br.DestEndpoint.IP,
		Verbose:    false,
	}); err != nil {
		return fmt.Errorf("error initializing AirPlay client: %w", err)
	}
	return nil
}

func (br *Bridge) initLocalDest() (err error) {
	if br.localAudioCtx == nil {
		br.localAudioCtx, err = oto.NewContext(44100, 2, 2, 12000)
		if err != nil {
			return fmt.Errorf("error creating new OTO context: %w", err)
		}
	}
	if br.localAudioDest != nil {
		_ = br.localAudioDest.Close()
	}
	br.localAudioDest = br.localAudioCtx.NewPlayer()
	return nil
}

func (br *Bridge) Wait() {
	<-br.done
}

func (br *Bridge) Stop() {
	br.done <- true
}
