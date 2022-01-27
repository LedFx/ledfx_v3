package audiobridge

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gordonklaus/portaudio"
	"github.com/hajimehoshi/oto"
	"io"
	"ledfx/integrations/airplay2"
	"ledfx/integrations/bluetooth"
	log "ledfx/logger"
	"time"
)

// NewBridge initializes a new bridge between a source and destination audio device.
func NewBridge(srcConfig EndpointConfig, dstConfig EndpointConfig, ledFxWriter io.Writer) (br *Bridge, err error) {
	br = &Bridge{
		SourceEndpoint:       &srcConfig,
		DestEndpoint:         &dstConfig,
		done:                 make(chan bool),
		ledFxWriter:          ledFxWriter,
		localAudioSourceDone: make(chan struct{}),
	}

	if err := br.initDevices(); err != nil {
		return nil, fmt.Errorf("error initializing devices: %w", err)
	}
	return br, nil
}

func (br *Bridge) initDevices() error {
	dstType := br.DestEndpoint.Type
	srcType := br.SourceEndpoint.Type

	if srcType == DeviceTypeBluetooth && dstType == DeviceTypeAirPlay {
		return ErrCannotBridgeBT2AP
	}

	switch dstType {
	case DeviceTypeAirPlay:
		// Do we have the required fields?
		if br.DestEndpoint.Name == "" {
			return errors.New("destination endpoint 'Name' field cannot be omitted if 'Type' is 'DeviceTypeAirPlay'")
		}

		// Connect to the AirPlay server
		if err := br.initAirPlayDest(); err != nil {
			return fmt.Errorf("error initializing AirPlay destination client: %w", err)
		}
	case DeviceTypeBluetooth:
		// Do we have the required fields?
		if br.DestEndpoint.Name == "" && br.DestEndpoint.Mac == "" {
			return errors.New("destination endpoint must have either 'Name' or 'Mac' specified")
		}

		// Connect to the Bluetooth device
		if err := br.initBluetoothDest(); err != nil {
			return fmt.Errorf("error initializing Bluetooth audio destination: %w", err)
		}

		// Bluetooth devices cannot simply be directly written to (AFAIK).
		// Instead, PulseAudio uses it as a playback device - meaning we can simply
		// use a local audio player to get the audio where it needs to go.
		if err := br.initLocalDest(); err != nil {
			return fmt.Errorf("error initializing local audio sink for Bluetooth: %w", err)
		}
	}

	switch srcType {
	case DeviceTypeAirPlay:
		// Do we have the required fields?
		if br.SourceEndpoint.Name == "" {
			return errors.New("source endpoint 'Name' field cannot be omitted if 'Type' is 'DeviceTypeAirPlay'")
		}

		// Spin up an AirPlay server
		if err := br.initAirPlaySource(); err != nil {
			return fmt.Errorf("error initializing AirPlay audio server: %w", err)
		}
	case DeviceTypeBluetooth:
		// Do we have the required fields?
		if br.SourceEndpoint.Name == "" {
			return errors.New("source endpoint 'Name' field cannot be omitted if 'Type' is 'DeviceTypeBluetooth'")
		}

		// Spin up a Bluetooth advertisement
		if err := br.initBluetoothSource(); err != nil {
			return fmt.Errorf("error initializing Bluetooth audio server: %w", err)
		}
	}

	switch {
	case srcType == DeviceTypeAirPlay && dstType == DeviceTypeAirPlay:
		br.airplayServer.SetClient(br.airplayClient)
		br.airplayServer.AddOutput(br.ledFxWriter)
	case srcType == DeviceTypeAirPlay && dstType == DeviceTypeBluetooth:
		br.airplayServer.AddOutput(br.localAudioDest)
		br.airplayServer.AddOutput(br.ledFxWriter)
	case srcType == DeviceTypeBluetooth && dstType == DeviceTypeBluetooth:
		// See annotation comment for br.InitLocalSource()
		if err := br.initLocalSource(br.ledFxWriter); err != nil {
			return fmt.Errorf("error initializing local PortAudio reader: %w", err)
		}

		// If dstType == DeviceTypeBluetooth then we use the default local audio destination.
		// The OS does this by default, so we only need to specify the ledFx writer.
	}

	return nil
}

func (br *Bridge) initAirPlaySource() (err error) {
	br.airplayServer = airplay2.NewServer(airplay2.Config{
		AdvertisementName: br.SourceEndpoint.Name,
		Port:              7000,
		VerboseLogging:    br.SourceEndpoint.Verbose,
	})
	if err = br.airplayServer.Start(); err != nil {
		return fmt.Errorf("error starting AirPlay server: %w", err)
	}
	return nil
}

func (br *Bridge) initAirPlayDest() (err error) {
	if br.airplayClient, err = airplay2.NewClient(airplay2.ClientDiscoveryParameters{
		DeviceNameRegex: br.DestEndpoint.Name,
		DeviceIP:        br.DestEndpoint.IP,
		Verbose:         br.DestEndpoint.Verbose,
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

// initLocalSource(): Under the hood, this function copies
// the local PortAudio stream to Bridge.ledFxWriter.
func (br *Bridge) initLocalSource(writeTo io.Writer) (err error) {
	if err = portaudio.Initialize(); err != nil {
		return fmt.Errorf("error initializing PortAudio: %w", err)
	}
	buffer := bytes.NewBuffer(make([]byte, 44100*2))
	if br.localAudioSource, err = portaudio.OpenDefaultStream(
		2,
		0,
		44100,
		buffer.Len(),
		func(in []float32) {
			for i := range in {
				buffer.WriteByte(byte(in[i]))
			}
			buffer.WriteTo(writeTo)
		},
	); err != nil {
		return fmt.Errorf("error opening default PortAudio stream: %w", err)
	}
	go func() {
		defer portaudio.Terminate()
		if err := br.localAudioSource.Start(); err != nil {
			log.Logger.Errorf("error starting local PortAudio stream as SOURCE: %v", err)
			return
		}
		defer br.localAudioSource.Close()
		<-br.localAudioSourceDone
	}()
	return nil
}

func (br *Bridge) initBluetoothDest() (err error) {
	if br.bluetoothClient != nil {
		br.bluetoothClient.Close()
	}
	if br.bluetoothClient, err = bluetooth.NewClient(); err != nil {
		return fmt.Errorf("error initializing Bluetooth client: %w", err)
	}
	if err := br.bluetoothClient.SearchAndConnect(bluetooth.SearchTargetConfig{
		DeviceRegex:          br.DestEndpoint.Name,
		ConnectRetryCoolDown: 250 * time.Millisecond,
		DeviceAddress:        br.DestEndpoint.Mac,
	}); err != nil {
		return fmt.Errorf("error querying for Bluetooth device: %w", err)
	}
	return nil
}

func (br *Bridge) initBluetoothSource() (err error) {
	if br.bluetoothServer != nil {
		br.bluetoothServer.CloseApp()
	}
	if br.bluetoothServer, err = bluetooth.NewServer(br.SourceEndpoint.Name); err != nil {
		return fmt.Errorf("error creating new Bluetooth server handler: %w", err)
	}
	if err = br.bluetoothServer.Serve(); err != nil {
		return fmt.Errorf("error serving bluetooth advertisement: %w", err)
	}
	return nil
}

func (br *Bridge) Wait() {
	<-br.done
}

func (br *Bridge) stop(notifyDone bool) {
	if notifyDone {
		defer func() {
			go func() {
				br.done <- true
			}()
		}()
	}

	switch br.SourceEndpoint.Type {
	case DeviceTypeAirPlay:
		br.airplayServer.Stop()
		log.Logger.Infof("Stopped AirPlay server")
	case DeviceTypeBluetooth:
		br.bluetoothServer.CloseApp()
		log.Logger.Infof("Stopped Bluetooth server")
	}

	switch br.DestEndpoint.Type {
	case DeviceTypeAirPlay:
		br.airplayClient.Close()
		log.Logger.Infof("Stopped AirPlay client")
	case DeviceTypeBluetooth:
		// Close the player instead of the OTO context.
		br.bluetoothClient.Close()
		log.Logger.Infof("Stopped Bluetooth client")
		br.localAudioDest.Close()
		log.Logger.Infof("Stopped OTO destination")
	}

	if br.SourceEndpoint.Type == DeviceTypeBluetooth && br.DestEndpoint.Type == DeviceTypeBluetooth {
		br.localAudioSourceDone <- struct{}{}
		log.Logger.Infof("Stopped local audio source")
	}

}

// Stop stops the bridge. Any further references to 'br *Bridge'
// may cause a runtime panic.
func (br *Bridge) Stop() {
	if br != nil {
		br.stop(true)
	}
}

// Reset resets the bridge and all active services with the newly
// provided configurations.
func (br *Bridge) Reset(newSourceConf, newDestConf EndpointConfig) error {
	if newSourceConf.Type == DeviceTypeBluetooth && newDestConf.Type == DeviceTypeAirPlay {
		return ErrCannotBridgeBT2AP
	}
	br.stop(false)
	br.DestEndpoint = &newDestConf
	br.SourceEndpoint = &newSourceConf
	if err := br.initDevices(); err != nil {
		return fmt.Errorf("error re-initializing devices: %w", err)
	}
	return nil
}
