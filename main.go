package main

import (
	"flag"
	"fmt"
	"ledfx/audio"
	"ledfx/audio/audiobridge"
	"ledfx/config"
	"ledfx/constants"
	"ledfx/effect"
	"ledfx/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"fyne.io/systray"
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
	err := config.Initialise()
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
	// if config.GlobalConfig.Version {
	// 	fmt.Println("LedFx " + constants.VERSION)
	// 	return
	// }

	// Print the cli logo
	// err := utils.PrintLogo()
	// if err != nil {
	// 	logger.Logger.Fatal(err)
	// }
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

	mux := http.DefaultServeMux
	a := audio.NewAnalyzer() // we only need one audio analyser. should it be more closely tied to the bridge?
	defer a.Cleanup()
	// if err := bridgeapi.NewServer(a.BufferCallback, mux); err != nil {
	// 	logger.Logger.WithField("category", "AudioBridge Server Init").Fatalf("Error initializing audio bridge server: %v", err)
	// }

	br, err := audiobridge.NewBridge(a.BufferCallback)
	if err != nil {
		log.Fatalf("Error initializing new bridge: %v\n", err)
	}
	defer br.Stop() // this sould also call analyser.cleanup(), it has to explicitly free some c memory

	audio.LogAudioDevices()
	audiodevice, err := audio.GetDeviceByID("9f012a5ef29af5e7b226bae734a8cb2ad229f063") // get from config
	if err != nil {
		log.Fatal(err)
	}
	if err := br.StartLocalInput(audiodevice, true); err != nil {
		log.Fatalf("Error starting local input: %v\n", err)
	}

	effect.NewAPI(mux)

	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		logger.Logger.WithField("category", "HTTP Listener").Fatalf("Error listening and serving: %v", err)
	}
}

func shutdown() {
	logger.Logger.Info("Shutting down LedFx")
	// kill systray
	systray.Quit()
}
