package utils

import (
	"fmt"
	"ledfx/api"
	"ledfx/audio"
	"ledfx/bridgeapi"
	log "ledfx/logger"
	"net/http"
	"path/filepath"
	"runtime"

	pretty "github.com/fatih/color"
	"github.com/rs/cors"
)

func ServeHttp() {
	DownloadFrontend()
	serveFrontend := http.FileServer(http.Dir("frontend"))
	api.HandleApi()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Logger.WithField("category", "HTTP Regexp").Debugf("Request asked for %s", r.URL.Path)
		if filepath.Ext(r.URL.Path) == "" {
			log.Logger.WithField("category", "HTTP Regexp").Debugln("Serving index.html")
			http.ServeFile(w, r, "frontend/index.html")
		} else {
			log.Logger.WithField("category", "HTTP Regexp").Debugf("Serving HTTP for path: %s", r.URL.Path)
			serveFrontend.ServeHTTP(w, r)
		}
	})
}

func InitFrontend(ip string, port int) {
	fxHandler, err := audio.NewFxHandler()
	if err != nil {
		log.Logger.Fatalf("Error initializing new FX handler: %v", err)
	}
	go func(fxh *audio.FxHandler) {
		err := bridgeapi.NewServer(fxh.Callback, http.DefaultServeMux)
		if err != nil {
			log.Logger.Fatal(err)
		}

		if err = http.ListenAndServe(fmt.Sprintf("%s:%d", ip, port), cors.AllowAll().Handler(http.DefaultServeMux)); err != nil {
			log.Logger.Fatal(err)
		}
	}(fxHandler)
	borderPrinter := pretty.New(pretty.BgBlack, pretty.FgRed)
	boldPrinter := pretty.New(pretty.BgBlack, pretty.FgRed, pretty.Bold)
	namePrinter := pretty.New(pretty.BgBlack, pretty.FgWhite, pretty.Faint)
	keyCombPrinter := pretty.New(pretty.BgBlack, pretty.FgHiYellow)
	linkPrinter := pretty.New(pretty.BgBlack, pretty.FgHiBlue, pretty.Bold, pretty.Underline)

	borderPrinter.Print("╭───────────────────────────────────────────────────────╮")
	fmt.Println()
	borderPrinter.Print("│                ")

	boldPrinter.Print("LedFX-Frontend ")
	namePrinter.Print("by Blade ")

	borderPrinter.Print("               │")
	fmt.Println()
	borderPrinter.Print("├───────────────────────────────────────────────────────┤")
	fmt.Println()
	borderPrinter.Print("│                                                       │")
	fmt.Println()
	borderPrinter.Print("│   ")

	switch runtime.GOOS {
	case "darwin":
		keyCombPrinter.Print("[CMD]+Click: ")
		linkPrinter.Print("http://localhost:8080/#/?newCore=1")
		borderPrinter.Print("     │")
	default:
		keyCombPrinter.Print("[CTRL]+Click: ")
		linkPrinter.Print("http://localhost:8080/#/?newCore=1")
		borderPrinter.Print("    │")
	}
	fmt.Println()
	borderPrinter.Print("│                                                       │")
	fmt.Println()
	borderPrinter.Print("╰───────────────────────────────────────────────────────╯")
	fmt.Println()
}
