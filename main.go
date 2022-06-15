package main

import (
	"fmt"
	"ledfx/audio"
	"ledfx/config"
	"ledfx/constants"
	"ledfx/effect"
	"ledfx/frontend"
	"ledfx/logger"
	"ledfx/util"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	pretty "github.com/fatih/color"
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
	coreConfig := config.GetCore()
	logger.Logger.SetLevel(logrus.Level(5 - coreConfig.LogLevel))
	hostport := fmt.Sprintf("%s:%d", coreConfig.Host, coreConfig.Port)
	url := fmt.Sprintf("http://%s", hostport)

	logger.Logger.Info("Info message logging enabled")
	logger.Logger.Debug("Debug message logging enabled")

	// Print the cli logo and welcome message
	if !coreConfig.NoLogo {
		util.PrintLogo()
	}
	fmt.Println()
	fmt.Println("Welcome to LedFx " + constants.VERSION)
	fmt.Println()

	// Print URL to go to
	linkPrinter := pretty.New(pretty.FgHiBlue, pretty.Bold, pretty.Underline)
	fmt.Println("Access LedFx through the web interface.")
	switch runtime.GOOS {
	case "darwin":
		fmt.Print("[CMD]+Click: ")
	default:
		fmt.Print("[CTRL]+Click: ")
	}
	linkPrinter.Print(url)
	fmt.Println()
	fmt.Println()

	frontend.Update()

	if coreConfig.OpenUi {
		util.OpenBrowser(url)
		logger.Logger.Info("Automatically opened the browser")
	}

	// TODO: handle other flags
	/**
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
	frontend.ServeHttp(mux)

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

	logger.Logger.WithField("category", "HTTP Listener").Infof("Starting LedFx HTTP Server at %s", hostport)
	if err := http.ListenAndServe(hostport, mux); err != nil {
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
