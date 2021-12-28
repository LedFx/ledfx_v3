package main

import (
	"fmt"
	"ledfx/api"
	"ledfx/color"
	"ledfx/config"
	"ledfx/constants"
	"ledfx/device"
	"ledfx/logger"
)

var Config config.Config

func init() {

	err := config.InitFlags()
	if err != nil {
		logger.Logger.Fatal(err)
	}

	conf, err := config.LoadConfig()
	if err != nil {
		logger.Logger.Fatal(err)
	}

	Config = conf

	_, err = logger.Init(Config)
	if err != nil {
		logger.Logger.Fatal(err)
	}

}

func main() {
	if Config.Version {
		fmt.Println("LedFx " + constants.VERSION)
		return
	}

	err := constants.PrintLogo()
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
	for _, d := range Config.Devices {
		if d.Id == "wled" {
			deviceConfig = d.Config
		}
	}

	// NOTE: This type of code should be run in a goroutine
	var device device.Device = &device.UdpDevice{
		Name:     "UDP",
		Port:     deviceConfig.Port,
		Protocol: device.UdpProtocols[deviceConfig.UdpPacketType],
	}

	data := [50]color.Color{}
	for i := range data {
		data[i], err = color.NewColor(color.LedFxColors["pink"])
		if err != nil {
			logger.Logger.Fatal(err)
		}
	}
	err = device.Init()
	if err != nil {
		logger.Logger.Fatal(err)
	}
	err = device.SendData(data[:])
	if err != nil {
		logger.Logger.Fatal(err)
	}

	defer device.Close()

	err = api.InitApi(Config.Port)
	if err != nil {
		logger.Logger.Fatal(err)
	}

}
