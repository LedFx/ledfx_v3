package controller

import (
	"fmt"

	"github.com/LedFx/ledfx/pkg/config"
	"github.com/LedFx/ledfx/pkg/device"
	"github.com/LedFx/ledfx/pkg/effect"
	"github.com/LedFx/ledfx/pkg/event"
	"github.com/LedFx/ledfx/pkg/logger"
)

// links effect IDs to controller IDs
var connectionsEffect = map[string]string{}

// links device IDs to controller IDs
var connectionsDevice = map[string]string{}

func LoadConnectionsFromConfig() {
	effects, devices := config.GetConnections()
	for eID, vID := range effects {
		err := ConnectEffect(eID, vID)
		if err != nil {
			logger.Logger.WithField("context", "Controller Connections").Fatal(err)
		}
	}
	for dID, vID := range devices {
		err := ConnectDevice(dID, vID)
		if err != nil {
			logger.Logger.WithField("context", "Controller Connections").Fatal(err)
		}
	}
	// invoke event
	event.Invoke(event.ConnectionsUpdate,
		map[string]interface{}{
			"effects": connectionsEffect,
			"devices": connectionsDevice,
		})
}

func ConnectEffect(effectID, controllerID string) error {
	// make sure effect exists
	e, err := effect.Get(effectID)
	if err != nil {
		return err
	}
	// make sure controller exists
	v, err := Get(controllerID)
	if err != nil {
		return err
	}
	// if already connected, don't continue
	if v.Effect != nil && v.Effect.ID == effectID {
		return nil
	}

	for eID, vID := range connectionsEffect {
		// -> controller can only have one effect.
		// if there's already an effect connected to the controller, disconnect it
		// -> effect can only output to one controller.
		// if it's already assigned to a controller, disconnect it first
		if eID == effectID || vID == controllerID {
			delete(connectionsEffect, eID)
			otherv, _ := Get(vID)
			otherv.Effect = nil
		}
	}
	connectionsEffect[effectID] = controllerID
	v.Effect = e
	// if the controller has a device, initialise the effect with the pixel count
	if len(v.Devices) != 0 {
		v.Effect.UpdatePixelCount(v.PixelCount())
	}
	config.SetConnections(connectionsEffect, connectionsDevice)
	// invoke event
	event.Invoke(event.ConnectionsUpdate,
		map[string]interface{}{
			"effects": connectionsEffect,
			"devices": connectionsDevice,
		})
	logger.Logger.WithField("context", "Controllers").Infof("Connected %s to %s", effectID, controllerID)
	return nil
}

func ConnectDevice(deviceID, controllerID string) error {
	// make sure device exists
	dev, err := device.Get(deviceID)
	if err != nil {
		return err
	}
	// make sure controller exists
	v, err := Get(controllerID)
	if err != nil {
		return err
	}
	// if already connected, don't continue
	for _, d := range v.Devices {
		if d.ID == deviceID {
			return nil
		}
	}
	connectionsDevice[deviceID] = controllerID
	v.Devices[dev.ID] = dev
	if dev.State != device.Connected {
		err = dev.Connect()
	}
	// if the controller has an effect, initialise it with the pixel count
	if v.Effect != nil {
		v.Effect.UpdatePixelCount(v.PixelCount())
	}
	config.SetConnections(connectionsEffect, connectionsDevice)
	// invoke event
	event.Invoke(event.ConnectionsUpdate,
		map[string]interface{}{
			"effects": connectionsEffect,
			"devices": connectionsDevice,
		})
	logger.Logger.WithField("context", "Controllers").Infof("Connected %s to %s", deviceID, controllerID)
	return err
}

func DisconnectEffect(effectID, controllerID string) error {
	// make sure controller exists
	_, err := Get(controllerID)
	if err != nil {
		return err
	}
	vID, connected := connectionsEffect[effectID]
	if !connected || controllerID != vID {
		err = fmt.Errorf("effect %s and controller %s are not connected", effectID, controllerID)
		return err
	}
	v, _ := Get(vID)
	if v.Effect == nil {
		return nil
	}
	delete(connectionsEffect, effectID)
	v.Stop()
	v.Effect = nil
	config.SetConnections(connectionsEffect, connectionsDevice)
	// invoke event
	event.Invoke(event.ConnectionsUpdate,
		map[string]interface{}{
			"effects": connectionsEffect,
			"devices": connectionsDevice,
		})
	logger.Logger.WithField("context", "Controllers").Infof("Disconnected %s from %s", effectID, controllerID)
	return err
}

func DisconnectDevice(deviceID, controllerID string) error {
	// make sure controller exists
	_, err := Get(controllerID)
	if err != nil {
		return err
	}
	// delete it from connections
	vID, connected := connectionsDevice[deviceID]
	if !connected || controllerID != vID {
		err = fmt.Errorf("device %s and controller %s are not connected", deviceID, controllerID)
		return err
	}
	delete(connectionsDevice, deviceID)
	// delete it from controller
	v, _ := Get(vID)
	d, exists := v.Devices[deviceID]
	if !exists {
		return nil
	}
	if d.State == device.Connected {
		err = d.Disconnect()
	}
	delete(v.Devices, deviceID)
	if len(v.Devices) == 0 {
		v.Stop()
	}
	config.SetConnections(connectionsEffect, connectionsDevice)
	// invoke event
	event.Invoke(event.ConnectionsUpdate,
		map[string]interface{}{
			"effects": connectionsEffect,
			"devices": connectionsDevice,
		})
	logger.Logger.WithField("context", "Controllers").Infof("Disconnected %s from %s", deviceID, controllerID)
	return err
}
