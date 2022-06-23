package virtual

import (
	"fmt"
	"ledfx/config"
	"ledfx/device"
	"ledfx/effect"
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
	// effect can only output to one virtual.
	// if it's already assigned to a virtual, disconnect it first
	for eID, vID := range connectionsEffect {
		if eID == effectID {
			delete(connectionsEffect, eID)
			otherv, _ := Get(vID)
			otherv.Effect = nil
		}
	}
	connectionsEffect[effectID] = virtualID
	v.Effect = e
	// if the virtual has a device, initialise the effect with the pixel count
	if v.Device != nil {
		v.Effect.UpdatePixelCount(v.Device.Config.PixelCount)
	}
	if !dontSave {
		config.SetConnections(connectionsEffect, connectionsDevice)
	}
	return nil
}

func ConnectDevice(deviceID, virtualID string) error {
	// make sure device exists
	d, err := device.Get(deviceID)
	if err != nil {
		return err
	}
	// make sure virtual exists
	v, err := Get(virtualID)
	if err != nil {
		return err
	}
	// if already connected, don't continue
	if v.Device != nil && v.Device.ID == deviceID {
		return nil
	}
	// virtual can only output to one device.
	// if it's already assigned to a virtual, disconnect it first
	for dID, vID := range connectionsDevice {
		if dID == deviceID {
			delete(connectionsDevice, dID)
			otherv, _ := Get(vID)
			if d.State == device.Connected {
				err = otherv.Device.Disconnect()
			}
			otherv.Device = nil
			if err != nil {
				return err
			}
		}
	}
	connectionsDevice[deviceID] = virtualID
	v.Device = d
	if d.State != device.Connected {
		err = d.Connect()
	}
	// if the virtual has an effect, initialise it with the pixel count
	if v.Effect != nil {
		v.Effect.UpdatePixelCount(d.Config.PixelCount)
	}
	if !dontSave {
		config.SetConnections(connectionsEffect, connectionsDevice)
	}
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
	v.Effect = nil
	if !dontSave {
		config.SetConnections(connectionsEffect, connectionsDevice)
	}
	return err
}

func DisconnectDevice(deviceID, virtualID string) error {
	// make sure virtual exists
	_, err := Get(virtualID)
	if err != nil {
		return err
	}
	vID, connected := connectionsDevice[deviceID]
	if !connected || virtualID != vID {
		err = fmt.Errorf("device %s and virtual %s are not connected", deviceID, virtualID)
		return err
	}
	v, _ := Get(vID)
	if v.Device == nil {
		return nil
	}
	delete(connectionsDevice, deviceID)
	if v.Device.State == device.Connected {
		err = v.Device.Disconnect()
	}
	v.Device = nil
	if !dontSave {
		config.SetConnections(connectionsEffect, connectionsDevice)
	}
	return err
}
