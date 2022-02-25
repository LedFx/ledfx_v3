package main

import (
	"flag"
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
	flag.StringVar(&ip, "ip", "0.0.0.0", "The IP address the frontend will run on")
	flag.IntVar(&port, "port", 8080, "The port the frontend will run on")

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

var (
	ip   string
	port int
)

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
	//go audio.CaptureDemo()

	go func() {
		utils.SetupRoutes()
	}()

	go func() {
		utils.InitFrontend(ip, port)
	}()

	go func() {
		err = utils.ScanZeroconf()
		if err != nil {
			logger.Logger.Fatal(err)
		}
	}()

	systray.Run(utils.OnReady, utils.OnExit)
	os.TempDir()

	err = virtual.LoadVirtuals()
	if err != nil {
		logger.Logger.Warn(err)
	}
}

func shutdown() {
	logger.Logger.Info("Shutting down LedFx")
	// kill systray
	systray.Quit()
}
