package effect

import (
	"fmt"
	"ledfx/color"
	"ledfx/config"
	"ledfx/device"
	"ledfx/logger"
	"ledfx/tickpool"
	"math"
	"time"
)

// Effect is the interface for an effect
type Effect interface {
	AssembleFrame(phase float64, ledCount int, effectColor color.Color) (colors []color.Color)
	Pixels []color.Color
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
func StartEffect(deviceConfig config.DeviceConfig, effect Effect, clr string, fps float64, done <-chan bool) error {
	usPerFrame := 1.0 / fps
	usPerFrameDuration := time.Duration(usPerFrame*1000000.0) * time.Microsecond
	logger.Logger.WithField("category", "Effect Starter").Debugf("[FPS=%v, FrameDuration=%vμs]", fps, usPerFrameDuration.Microseconds())

	ticker := tickpool.Get(usPerFrameDuration)
	defer tickpool.Put(ticker)

	phase := 0.0 // phase of the effect (range 0.0 to 2π)

	// TODO: choose type of dev dynamically based on the deviceConfig
	dev := &device.UDPDevice{
		Name:     deviceConfig.Name,
		Port:     deviceConfig.Port,
		Protocol: device.UDPProtocols[deviceConfig.UdpPacketType],
		Config:   deviceConfig,
	}

	if err := dev.Init(); err != nil {
		return fmt.Errorf("error initializing device while starting effect: %w", err)
	}
	defer dev.Close()

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

			if err = dev.SendData(effect.AssembleFrame(phase, dev.Config.PixelCount, newColor), 0x00); err != nil {
				return fmt.Errorf("error sending data assembled frame data on <-done notify: %w", err)
			}
			return nil
		case <-ticker.C:
			// TODO: get pixelCount and color from config
			// TODO: this should be
			newColor, err := color.NewColor(clr)
			if err != nil {
				return err
			}

			if err = dev.SendData(effect.AssembleFrame(phase, dev.Config.PixelCount, newColor), 0xff); err != nil {
				return fmt.Errorf("error sending data assembled from frame data on <-ticker.C notify: %w", err)
			}

			// Increment the phase (range: 0 - 2π)
			phase += ((2 * math.Pi) / fps) * speed
			if phase >= (2 * math.Pi) {
				phase = 0.0
			}
		}
	}
}
func StopEffect(deviceConfig config.DeviceConfig, effect Effect, clr string, fps int, done <-chan bool) error {
	// TODO: choose type of dev dynamically based on the deviceConfig
	dev := &device.UDPDevice{
		Name:     deviceConfig.Name,
		Port:     deviceConfig.Port,
		Protocol: device.UDPProtocols[deviceConfig.UdpPacketType],
		Config:   deviceConfig,
	}

	if err := dev.Init(); err != nil {
		return fmt.Errorf("error initializing device while stopping effect: %w", err)
	}
	defer dev.Close()

	fmt.Println("Done!")

	newColor, err := color.NewColor("#000000")
	if err != nil {
		return err
	}

	if err = dev.SendData(effect.AssembleFrame(0.0, dev.Config.PixelCount, newColor), 0x00); err != nil {
		return fmt.Errorf("error sending data to dev: %w", err)
	}

	return nil
}

// TODO: StopEffect
