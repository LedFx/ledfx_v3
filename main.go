package main

import (
	"fmt"
	"ledfx/audio"
	"ledfx/bridgeapi"
	"ledfx/config"
	"ledfx/constants"
	"ledfx/device"
	"ledfx/effect"
	"ledfx/frontend"
	"ledfx/logger"
	"ledfx/util"
	"ledfx/virtual"
	"ledfx/websocket"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"fyne.io/systray"
	pretty "github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

var wg sync.WaitGroup

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
	settings := config.GetSettings()
	logger.Logger.SetLevel(logrus.Level(5 - settings.LogLevel))
	hostport := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	url := fmt.Sprintf("http://%s", hostport)
	logger.Logger.Info("Info message logging enabled")
	logger.Logger.Debug("Debug message logging enabled")

	// Check for updates
	if settings.NoUpdate {
		logger.Logger.Warn("Not checking for updates")
	} else {
		frontend.Update()
	}

	// run systray
	if settings.NoTray {
		logger.Logger.Warn("Not creating system tray icon")
	} else {
		go systray.Run(util.StartTray(url), util.StopTray)
	}

	// TODO: handle other flags
	/**
	  profiler
	*/
	// profiler.Start()

	//go audio.CaptureDemo()

	// Set up API routes
	// Initialise frontend
	// Scan for WLED
	// Run systray
	// load effects, devices, virtuals

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
	// 	logger.Logger.WithField("context", "HTTP Listener").Fatalf("Error starting local input: %v\n", err)
	// }

	err := effect.LoadFromConfig()
	if err != nil {
		logger.Logger.WithField("context", "Load Effects from Config").Fatal(err)
	}
	err = device.LoadFromConfig()
	if err != nil {
		logger.Logger.WithField("context", "Load Devices from Config").Fatal(err)
	}
	err = virtual.LoadFromConfig()
	if err != nil {
		logger.Logger.WithField("context", "Load Virtuals from Config").Fatal(err)
	}
	virtual.LoadConnectionsFromConfig()
	virtual.LoadStatesFromConfig()

	// Handle WLED scanning
	if !settings.NoScan {
		util.EnableScan()
	} else {
		logger.Logger.Warning("WLED scanning is disabled")
	}

	// Add routes
	mux := http.DefaultServeMux
	effect.NewAPI(mux)
	device.NewAPI(mux)
	virtual.NewAPI(mux)
	config.NewAPI(mux)
	frontend.NewServer(mux)
	websocket.Serve(mux)
	bridgeServer, err := bridgeapi.NewServer(audio.Analyzer.BufferCallback, mux)

	// Start audio bridge
	if err != nil {
		logger.Logger.WithField("context", "AudioBridge").Fatalf("Error initializing AudioBridge server: %v", err)
	} else {
		logger.Logger.WithField("context", "AudioBridge").Info("Initialised AudioBridge server")
	}
	defer bridgeServer.Br.Stop()
	if err := bridgeServer.Br.StartLocalInput(config.GetLocalInput()); err != nil {
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error starting local input: %v\n", err)
	}

	// Start web server
	wg.Add(1)
	logger.Logger.WithField("context", "HTTP Listener").Infof("Starting LedFx HTTP Server at %s", hostport)
	go func() {
		defer wg.Done()
		if err := http.ListenAndServe(hostport, mux); err != nil {
			logger.Logger.WithField("context", "HTTP Listener").Fatalf("Error listening and serving: %v", err)
		}
	}()

	// Print the cli logo and welcome message
	if !settings.NoLogo {
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

	// Open the browser
	if settings.OpenUi {
		util.OpenBrowser(url)
		logger.Logger.Info("Automatically opened the browser")
	}

	// Wait for all running goroutines to finish
	wg.Wait()
}

func shutdown() {
	logger.Logger.Info("Shutting down LedFx")
	// kill analyzer
	audio.Analyzer.Cleanup()

	// kill systray
	systray.Quit()
}
