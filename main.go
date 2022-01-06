package main

import (
	"fmt"
	"ledfx/audio"
	"ledfx/color"
	"ledfx/config"
	"ledfx/constants"
	"ledfx/device"
	"ledfx/logger"
	"ledfx/utils"
	"log"

	"github.com/getlantern/systray"
)

func init() {
	// Initialize Config
	err := config.InitConfig()
	if err != nil {
		log.Println(err)
	}

	// Initialize Logger
	_, err = logger.Init(config.GlobalConfig)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	// Just print version and return if flag is set
	if config.GlobalConfig.Version {
		fmt.Println("LedFx " + constants.VERSION)
		return
	}

	// Print the cli logo
	err := utils.PrintLogo()
	if err != nil {
		logger.Logger.Fatal(err)
	}
	fmt.Println("Welcome to LedFx " + constants.VERSION)
	fmt.Println()

	logger.Logger.Info("Verbose logging enabled")
	logger.Logger.Debug("Very verbose logging enabled")

	// TODO: handle other flags
	/**
	  OpenUi
	  Host
	  Offline
	  SentryCrash
	*/

	// REMOVEME: testing only
	// Initialize config
	var deviceConfig config.DeviceConfig
	var foundDevice bool = false
	for _, d := range config.GlobalConfig.Devices {
		if d.Type == "udp" {
			deviceConfig = d.Config
			foundDevice = true
			break
		}
	}

	if !foundDevice {
		logger.Logger.Warn("No UDP device found in config")
	} else {

		// NOTE: This type of code should be run in a goroutine
		var device = &device.UdpDevice{
			Name:     deviceConfig.Name,
			Port:     deviceConfig.Port,
			Protocol: device.UdpProtocols[deviceConfig.UdpPacketType],
			Config:   deviceConfig,
		}

		data := []color.Color{}
		for i := 0; i < device.Config.PixelCount; i++ {
			newColor, err := color.NewColor(color.LedFxColors["orange"])
			data = append(data, newColor)
			if err != nil {
				logger.Logger.Fatal(err)
			}
		}
		err = device.Init()
		if err != nil {
			logger.Logger.Fatal(err)
		}
		err = device.SendData(data, 0x01)
		if err != nil {
			logger.Logger.Fatal(err)
		}

		defer device.Close()
	}
	// REMOVEME: END

	audio.Enumerate()

	go func() {
		utils.SetupRoutes()
	}()

	go func() {
		utils.InitFrontend()
	}()

	go func() {
		err = utils.ScanZeroconf()
		if err != nil {
			logger.Logger.Fatal(err)
		}
	}()

	systray.Run(utils.OnReady, nil)

}
