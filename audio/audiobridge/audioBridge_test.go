package audiobridge

import (
	"github.com/pkg/errors"
	"io"
	log "ledfx/logger"
	"testing"
	"time"
)

// NOTE: All the tests below will only succeed if the following instructions are followed:
// 1). Must be run on a Raspberry Pi with Bluez
// 2). 'dstConfAP.Name' must be changed to a regex pattern that matches an AirPlay advertisement within range of the Pi.
// 3). 'dstConfBT.Name' must be changed to a regex pattern that matches a Bluetooth advertisement within range of the Pi.
func TestAudioBridge_Reset(t *testing.T) {
	srcConfAP := EndpointConfig{
		Type:    DeviceTypeAirPlay,
		Name:    "LedFX-Input-Test",
		Verbose: true,
	}
	srcConfBT := EndpointConfig{
		Type:    DeviceTypeBluetooth,
		Name:    "LedFX-Input-Test",
		Verbose: true,
	}

	dstConfAP := EndpointConfig{
		Type:    DeviceTypeAirPlay,
		Name:    "(?i)Output$",
		Verbose: true,
	}
	dstConfBT := EndpointConfig{
		Type:    DeviceTypeBluetooth,
		Name:    "(?i)K850$",
		Verbose: true,
	}

	bridge, err := NewBridge(srcConfAP, dstConfAP, io.Discard)
	if err != nil {
		t.Fatalf("Error initializing new bridge (AP -> AP): %v\n", err)
	}

	time.Sleep(10 * time.Second)
	log.Logger.WithField("component", "AP -> AP").Infoln("Success!")

	if err := bridge.Reset(srcConfAP, dstConfBT); err != nil {
		t.Fatalf("Error resetting bridge (AP -> BT): %v", err)
	}

	time.Sleep(10 * time.Second)
	log.Logger.WithField("component", "AP -> BT").Infoln("Success!")

	if err := bridge.Reset(srcConfBT, dstConfBT); err != nil {
		t.Fatalf("Error resetting bridge (BT -> BT): %v", err)
	}

	time.Sleep(10 * time.Second)
	log.Logger.WithField("component", "BT -> BT").Infoln("Success!")

	if err := bridge.Reset(srcConfBT, dstConfAP); !errors.Is(err, ErrCannotBridgeBT2AP) {
		t.Fatalf("Invalid error returned (BT -> AP, expected 'ErrCannotBridgeBT2AP'): %v\n", err)
	}

	log.Logger.WithField("component", "BT -> AP").Infoln("Success! (Proper error returned)")
	bridge.Stop()
	log.Logger.WithField("component", "Kill Bridge").Infoln("Success!")

}

func TestAudioBridge_Ap2Ap(t *testing.T) {
	// srcConf is the config for the endpoint from which
	// audio will be ingested, processed, converted, and
	// redistributed.
	srcConf := EndpointConfig{
		Type:    DeviceTypeAirPlay,  // Spin up an AirPlay server
		Name:    "LedFX-Input-Test", // It will be advertised as "LedFX-Input-Test"
		Verbose: true,               // Enable verbosity because this is a test package.
	}

	dstConf := EndpointConfig{
		Type: DeviceTypeAirPlay, // Connect to an airplay server

		// We will connect to any AirPlay server that contains the string "Output" (case-insensitive)
		Name: "(?i)Output$", // Regular expressions are required here.

		Verbose: true, // Enable verbosity because this is a test package.
	}

	// Initialize the bridge. This will start everything, too.
	bridge, err := NewBridge(srcConf, dstConf, io.Discard)
	if err != nil {
		t.Fatalf("error creating new audio bridge: %v\n", err)
	}

	// Wait until the bridge loop stops. This can be called with Stop()
	bridge.Wait()

}

func TestAudioBridge_Ap2Bt(t *testing.T) {
	srcConf := EndpointConfig{
		Type:    DeviceTypeAirPlay,
		Name:    "LedFX-Input-Test",
		Verbose: true,
	}

	dstConf := EndpointConfig{
		Type: DeviceTypeBluetooth,

		Name: "(?i)k850$",

		Verbose: true,
	}

	bridge, err := NewBridge(srcConf, dstConf, io.Discard)
	if err != nil {
		t.Fatalf("error creating new audio bridge: %v\n", err)
	}

	bridge.Wait()

}
