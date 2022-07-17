package virtual

import (
	"fmt"
	"ledfx/config"
	"ledfx/device"
	"ledfx/effect"
	"ledfx/event"
	"ledfx/logger"
)

// links effect IDs to virtual IDs
var connectionsEffect = map[string]string{}

// links device IDs to virtual IDs
var connectionsDevice = map[string]string{}

var dontSave bool = false

func LoadConnectionsFromConfig() {
	dontSave = true
	defer func() { dontSave = false }()
	effects, devices := config.GetConnections()
	for eID, vID := range effects {
		err := ConnectEffect(eID, vID)
		if err != nil {
			logger.Logger.WithField("context", "Virtual Connections").Fatal(err)
		}
	}
	for dID, vID := range devices {
		err := ConnectDevice(dID, vID)
		if err != nil {
			logger.Logger.WithField("context", "Virtual Connections").Fatal(err)
		}
	}
	// invoke event
	event.Invoke(event.ConnectionsUpdate,
		map[string]interface{}{
			"effects": connectionsEffect,
			"devices": connectionsDevice,
		})
}

func ConnectEffect(effectID, virtualID string) error {
	// make sure effect exists
	e, err := effect.Get(effectID)
	if err != nil {
		return err
	}
	// make sure virtual exists
	v, err := Get(virtualID)
	if err != nil {
		return err
	}
	// if already connected, don't continue
	if v.Effect != nil && v.Effect.ID == effectID {
		return nil
	}

	for eID, vID := range connectionsEffect {
		// -> virtual can only have one effect.
		// if there's already an effect connected to the virtual, disconnect it
		// -> effect can only output to one virtual.
		// if it's already assigned to a virtual, disconnect it first
		if eID == effectID || vID == virtualID {
			delete(connectionsEffect, eID)
			otherv, _ := Get(vID)
			otherv.Effect = nil
		}
	}
	connectionsEffect[effectID] = virtualID
	v.Effect = e
	// if the virtual has a device, initialise the effect with the pixel count
	if len(v.Devices) != 0 {
		v.Effect.UpdatePixelCount(v.PixelCount())
	}
	if !dontSave {
		config.SetConnections(connectionsEffect, connectionsDevice)
		// invoke event
		event.Invoke(event.ConnectionsUpdate,
			map[string]interface{}{
				"effects": connectionsEffect,
				"devices": connectionsDevice,
			})
	}
	logger.Logger.WithField("context", "Virtuals").Infof("Connected %s to %s", effectID, virtualID)
	return nil
}

func ConnectDevice(deviceID, virtualID string) error {
	// make sure device exists
	dev, err := device.Get(deviceID)
	if err != nil {
		return err
	}
	// make sure virtual exists
	v, err := Get(virtualID)
	if err != nil {
		return err
	}
	// if already connected, don't continue
	for _, d := range v.Devices {
		if d.ID == deviceID {
			return nil
		}
	}
	connectionsDevice[deviceID] = virtualID
	v.Devices[dev.ID] = dev
	if dev.State != device.Connected {
		err = dev.Connect()
	}
	// if the virtual has an effect, initialise it with the pixel count
	if v.Effect != nil {
		v.Effect.UpdatePixelCount(v.PixelCount())
	}
	if !dontSave {
		config.SetConnections(connectionsEffect, connectionsDevice)
		// invoke event
		event.Invoke(event.ConnectionsUpdate,
			map[string]interface{}{
				"effects": connectionsEffect,
				"devices": connectionsDevice,
			})
	}
	logger.Logger.WithField("context", "Virtuals").Infof("Connected %s to %s", deviceID, virtualID)
	return err
}

func DisconnectEffect(effectID, virtualID string) error {
	// make sure virtual exists
	_, err := Get(virtualID)
	if err != nil {
		return err
	}
	vID, connected := connectionsEffect[effectID]
	if !connected || virtualID != vID {
		err = fmt.Errorf("effect %s and virtual %s are not connected", effectID, virtualID)
		return err
	}
	v, _ := Get(vID)
	if v.Effect == nil {
		return nil
	}
	delete(connectionsEffect, effectID)
	v.Stop()
	if !dontSave {
		config.SetConnections(connectionsEffect, connectionsDevice)
		// invoke event
		event.Invoke(event.ConnectionsUpdate,
			map[string]interface{}{
				"effects": connectionsEffect,
				"devices": connectionsDevice,
			})
	}
	logger.Logger.WithField("context", "Virtuals").Infof("Disconnected %s from %s", effectID, virtualID)
	return err
}

func DisconnectDevice(deviceID, virtualID string) error {
	// make sure virtual exists
	_, err := Get(virtualID)
	if err != nil {
		return err
	}
	// delete it from connections
	vID, connected := connectionsDevice[deviceID]
	if !connected || virtualID != vID {
		err = fmt.Errorf("device %s and virtual %s are not connected", deviceID, virtualID)
		return err
	}
	delete(connectionsDevice, deviceID)
	// delete it from virtual
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
	if !dontSave {
		config.SetConnections(connectionsEffect, connectionsDevice)
		// invoke event
		event.Invoke(event.ConnectionsUpdate,
			map[string]interface{}{
				"effects": connectionsEffect,
				"devices": connectionsDevice,
			})
	}
	logger.Logger.WithField("context", "Virtuals").Infof("Disconnected %s from %s", deviceID, virtualID)
	return err
}
