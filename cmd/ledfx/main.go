package main

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" //nolint:gosec
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/LedFx/ledfx/pkg/audio"
	"github.com/LedFx/ledfx/pkg/bridgeapi"
	"github.com/LedFx/ledfx/pkg/color"
	"github.com/LedFx/ledfx/pkg/config"
	"github.com/LedFx/ledfx/pkg/constants"
	"github.com/LedFx/ledfx/pkg/controller"
	"github.com/LedFx/ledfx/pkg/device"
	"github.com/LedFx/ledfx/pkg/effect"
	"github.com/LedFx/ledfx/pkg/event"
	"github.com/LedFx/ledfx/pkg/frontend"
	"github.com/LedFx/ledfx/pkg/logger"
	"github.com/LedFx/ledfx/pkg/util"
	"github.com/LedFx/ledfx/pkg/websocket"

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
		event.Invoke(event.Shutdown, map[string]interface{}{})
	}()
}

func main() {
	settings := config.GetSettings()
	logger.Logger.SetLevel(logrus.Level(5 - settings.LogLevel))
	hostport := net.JoinHostPort(settings.Host, fmt.Sprint(settings.Port))
	url := fmt.Sprintf("http://%s", hostport)
	logger.Logger.Info("Info message logging enabled")
	logger.Logger.Debug("Debug message logging enabled")

	// subscribe to shutdown event
	event.Subscribe(event.Shutdown, func(e *event.Event) { shutdown() })

	// Check for updates
	if settings.NoUpdate {
		logger.Logger.Warn("Not checking for updates")
	} else {
		frontend.Update()
	}

	// TODO: handle profiler flags
	config.AllowSaving = false
	err := effect.LoadFromConfig()
	if err != nil {
		logger.Logger.WithField("context", "Load Effects from Config").Fatal(err)
	}
	err = device.LoadFromConfig()
	if err != nil {
		logger.Logger.WithField("context", "Load Devices from Config").Fatal(err)
	}
	err = controller.LoadFromConfig()
	if err != nil {
		logger.Logger.WithField("context", "Load Controllers from Config").Fatal(err)
	}
	controller.LoadConnectionsFromConfig()
	controller.LoadStatesFromConfig()
	config.AllowSaving = true

	// Handle WLED scanning
	if !settings.NoScan {
		device.EnableScan()
	} else {
		logger.Logger.Warning("WLED scanning is disabled")
	}

	// Add routes
	mux := http.DefaultServeMux
	effect.NewAPI(mux)
	device.NewAPI(mux)
	controller.NewAPI(mux)
	config.NewAPI(mux)
	color.NewAPI(mux)
	frontend.NewServer(mux)
	websocket.Serve(mux)
	bridgeServer, err := bridgeapi.NewServer(audio.Analyzer.BufferCallback, mux)
	// Start audio bridge
	if err != nil {
		logger.Logger.WithField("context", "AudioBridge").Fatalf("Error initializing AudioBridge server: %v", err)
	} else {
		defer bridgeServer.Br.Stop()
		logger.Logger.WithField("context", "AudioBridge").Info("Initialised AudioBridge server")
	}

	if err := bridgeServer.Br.StartLocalInput(config.GetLocalInput()); err != nil {
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error starting local input: %v\n", err)
	}

	// Start web server
	wg.Add(1)
	logger.Logger.WithField("context", "HTTP Listener").Infof("Starting LedFx HTTP Server at %s", hostport)
	go func() {
		defer wg.Done()
		if err := http.ListenAndServe(hostport, setHeaders(mux)); err != nil {
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

	// run systray
	if settings.NoTray {
		logger.Logger.Warn("Not creating system tray icon")
	} else {
		systray.Run(StartTray(url), StopTray)
	}

	// Wait for all running goroutines to finish
	wg.Wait()
}

func shutdown() {
	logger.Logger.WithField("context", "Shutdown Handler").Info("Shutting down LedFx")

	logger.Logger.WithField("context", "Shutdown Handler").Info("Cleaning up audio analyzer")
	// kill analyzer
	audio.Analyzer.Cleanup()

	// kill systray
	if !config.GetSettings().NoTray {
		logger.Logger.WithField("context", "Shutdown Handler").Info("Shutting down Systray")
		systray.Quit()
	}
	os.Exit(0)
}

func setHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//anyone can make a CORS request (not recommended in production)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		//only allow GET, PUT, POST, DELETE and OPTIONS
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, POST, DELETE, OPTIONS")
		//Since I was building a REST API that returned JSON, I set the content type to JSON here.
		w.Header().Set("Content-Type", "application/json")
		//Allow requests to have the following headers
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, cache-control")
		//if it's just an OPTIONS request, nothing other than the headers in the response is needed.
		//This is essential because you don't need to handle the OPTIONS requests in your handlers now
		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)
	})
}
