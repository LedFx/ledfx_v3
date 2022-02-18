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
	go func() {
		err := bridgeapi.NewServer(func(buf audio.Buffer) {
			// No callback for now
		}, http.DefaultServeMux)
		if err != nil {
			log.Logger.Fatal(err)
		}

		if err = http.ListenAndServe(fmt.Sprintf("%s:%d", ip, port), cors.AllowAll().Handler(http.DefaultServeMux)); err != nil {
			log.Logger.Fatal(err)
		}
	}()
	pretty.Set(pretty.BgBlack, pretty.FgRed).Print("╭───────────────────────────────────────────────────────╮\n│               ")
	pretty.Set(pretty.BgBlack, pretty.FgRed, pretty.Bold).Print(" LedFx-Frontend")
	pretty.Set(pretty.BgBlack, pretty.FgWhite, pretty.Faint).Print(" by Blade ")
	pretty.Set(pretty.BgBlack, pretty.FgRed).Print("               │\n├───────────────────────────────────────────────────────┤\n│                                                       │\n│   ")
	switch runtime.GOOS {
	case "darwin":
		pretty.Set(pretty.BgBlack, pretty.FgHiYellow).Print("[CMD]+LMB: ")
	default:
		pretty.Set(pretty.BgBlack, pretty.FgHiYellow).Print("[CTRL]+Click: ")
	}
	pretty.Set(pretty.BgBlack, pretty.FgHiBlue, pretty.Bold, pretty.Underline).Print("http://localhost:8080/#/?newCore=1")
	switch runtime.GOOS {
	case "darwin":
		pretty.Set(pretty.BgBlack, pretty.FgRed).Print("       │\n")

	default:
		pretty.Set(pretty.BgBlack, pretty.FgRed).Print("    │\n")

	}
	pretty.Set(pretty.BgBlack, pretty.FgRed).Print("│                                                       │\n╰───────────────────────────────────────────────────────╯\n")
	pretty.Unset()

}
