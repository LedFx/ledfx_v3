package main

import (
	"fmt"
	"ledfx/audio"
	"ledfx/config"
	"ledfx/constants"
	"ledfx/effect"
	"ledfx/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
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

}

func main() {
	// Print the cli logo
	// err := utils.PrintLogo()
	// if err != nil {
	// 	logger.Logger.Fatal(err)
	// }
	fmt.Println()
	fmt.Println("Welcome to LedFx " + constants.VERSION)
	fmt.Println()

	coreConfig := config.GetCore()
	logger.Logger.SetLevel(logrus.Level(5 - coreConfig.LogLevel))

	logger.Logger.Info("Info message logging enabled")
	logger.Logger.Debug("Debug message logging enabled")

	// TODO: handle other flags
	/**
	  OpenUi
	  Host
	  Offline
	  SentryCrash
	  profiler
	*/
	// profiler.Start()

	//go audio.CaptureDemo()

	// Set up API routes
	// Initialise frontend
	// Scan for WLED
	// Run systray
	// load effects, devices, virtuals

	// run systray
	// systray.Run(utils.OnReady, utils.OnExit)

	err := effect.LoadFromConfig()
	if err != nil {
		logger.Logger.WithField("context", "Load Effects from Config").Fatal(err)
	}

	mux := http.DefaultServeMux
	// if err := bridgeapi.NewServer(a.BufferCallback, mux); err != nil {
	// 	logger.Logger.WithField("category", "AudioBridge Server Init").Fatalf("Error initializing audio bridge server: %v", err)
	// }

	// br, err := audiobridge.NewBridge(audio.Analyzer.BufferCallback)
	// if err != nil {
	// 	log.Fatalf("Error initializing new bridge: %v\n", err)
	// }
	// defer br.Stop()

	// // audio.LogAudioDevices()
	// audiodevice, err := audio.GetDeviceByID("9f012a5ef29af5e7b226bae734a8cb2ad229f063") // get from config
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if err := br.StartLocalInput(audiodevice, true); err != nil {
	// 	logger.Logger.WithField("category", "HTTP Listener").Fatalf("Error starting local input: %v\n", err)
	// }

	effect.NewAPI(mux)

	address := fmt.Sprintf("%s:%d", coreConfig.Host, coreConfig.Port)
	logger.Logger.WithField("category", "HTTP Listener").Infof("Starting LedFx HTTP Server at %s", address)
	if err := http.ListenAndServe(address, mux); err != nil {
		logger.Logger.WithField("category", "HTTP Listener").Fatalf("Error listening and serving: %v", err)
	}
}

func shutdown() {
	logger.Logger.Info("Shutting down LedFx")
	// kill analyzer
	audio.Analyzer.Cleanup()

	// kill systray
	// systray.Quit()
}
