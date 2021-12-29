package main

import (
	"fmt"
	"ledfx/api"
	"ledfx/color"
	"ledfx/config"
	"ledfx/constants"
	"ledfx/device"
	"ledfx/logger"
	"ledfx/utils"
)

func init() {

	err := config.InitFlags()
	if err != nil {
		logger.Logger.Fatal(err)
	}

	err = config.LoadConfig()
	if err != nil {
		logger.Logger.Fatal(err)
	}

	_, err = logger.Init(config.GlobalConfig)
	if err != nil {
		logger.Logger.Fatal(err)
	}

}

func main() {
	if config.GlobalConfig.Version {
		fmt.Println("LedFx " + constants.VERSION)
		return
	}

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
		return
	}

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
	err = device.SendData(data)
	if err != nil {
		logger.Logger.Fatal(err)
	}

	defer device.Close()

	err = api.InitApi(config.GlobalConfig.Port)
	if err != nil {
		logger.Logger.Fatal(err)
	}

}
