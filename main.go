package main

import (
	"fmt"
	"ledfx/audio"
	"ledfx/config"
	"ledfx/constants"
	"ledfx/logger"
	"ledfx/utils"
	"ledfx/virtual"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/getlantern/systray"
)

func init() {
	// Capture ctrl-c or sigterm to gracefully shutdown
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		shutdown()
		os.Exit(1)
	}()

	// Initialize Config
	err := config.InitConfig()
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

	audio.LogAudioDevices()
	go audio.CaptureDemo()

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

	systray.Run(utils.OnReady, utils.OnExit)

	err = virtual.LoadVirtuals()
	if err != nil {
		logger.Logger.Warn(err)
	}
}

func shutdown() {
	logger.Logger.Info("Shutting down LedFx")
	// kill systray
	utils.OnExit()
}
