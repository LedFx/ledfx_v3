package effect

import (
	"fmt"
	"ledfx/color"
	"ledfx/config"
	"ledfx/device"
	"ledfx/logger"
	"math"
	"time"
)

// Effect is the interface for an effect
type Effect interface {
	AssembleFrame(phase float64, ledCount int, effectColor color.Color) (colors []color.Color)
}

// Config is the configuration for an effect
type Config struct {
	Blur       float64
	Flip       bool
	Mirror     bool
	Brightness float32
	Background color.Color
}

// StartEffect starts a specific effect on a device at a given FPS
func StartEffect(deviceConfig config.DeviceConfig, effect Effect, clr string, fps int, done <-chan bool) error {
	logger.Logger.Debug(fmt.Sprintf("fps: %v", fps))
	usPerFrame := (float64(1.0) / float64(fps))
	usPerFrameDuration := time.Duration(usPerFrame*1000000.0) * time.Microsecond
	logger.Logger.Debug(fmt.Sprintf("usPerFrameDuration: %v", usPerFrameDuration.Microseconds()))
	ticker := time.NewTicker(usPerFrameDuration)
	phase := 0.0 // phase of the effect (range 0.0 to 2π)

	// TODO: choose type of device dynamically based on the deviceConfig
	var device = &device.UDPDevice{
		Name:     deviceConfig.Name,
		Port:     deviceConfig.Port,
		Protocol: device.UDPProtocols[deviceConfig.UdpPacketType],
		Config:   deviceConfig,
	}

	err := device.Init()
	if err != nil {
		logger.Logger.Fatal(err)
	}

	defer ticker.Stop()

	// TODO: this should be in effect config
	speed := 1.0 // beats per minute
	if clr == "" {
		clr = "#000fff"
	}

	for {
		select {
		case <-done:
			fmt.Println("Done!")

			newColor, err := color.NewColor(clr)
			if err != nil {
				return err
			}
			err = device.SendData(effect.AssembleFrame(phase, device.Config.PixelCount, newColor), 0x00)
			if err != nil {
				return err
			}
			device.Close()
			return nil
		case <-ticker.C:
			// TODO: get pixelCount and color from config
			// TODO: this should be
			newColor, err := color.NewColor(clr)
			if err != nil {
				return err
			}
			err = device.SendData(effect.AssembleFrame(phase, device.Config.PixelCount, newColor), 0xff)
			if err != nil {
				return err
			}
			// Increment the phase (range: 0 - 2π)
			phase += ((2 * math.Pi) / float64(fps)) * speed
			if phase >= (2 * math.Pi) {
				phase = 0.0
			}
		}
	}
}
func StopEffect(deviceConfig config.DeviceConfig, effect Effect, clr string, fps int, done <-chan bool) error {

	// TODO: choose type of device dynamically based on the deviceConfig
	var device = &device.UDPDevice{
		Name:     deviceConfig.Name,
		Port:     deviceConfig.Port,
		Protocol: device.UDPProtocols[deviceConfig.UdpPacketType],
		Config:   deviceConfig,
	}

	err := device.Init()
	if err != nil {
		logger.Logger.Fatal(err)
	}

	fmt.Println("Done!")

	newColor, err := color.NewColor("#000000")
	if err != nil {
		return err
	}
	err = device.SendData(effect.AssembleFrame(0.0, device.Config.PixelCount, newColor), 0x00)
	if err != nil {
		return err
	}
	device.Close()
	return nil
}

// TODO: StopEffect
