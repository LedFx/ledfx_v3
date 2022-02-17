package utils

import (
	"fmt"
	pretty "github.com/fatih/color"
	"github.com/rs/cors"
	"ledfx/api"
	"ledfx/audio"
	"ledfx/bridgeapi"
	log "ledfx/logger"
	"net/http"
	"regexp"
	"runtime"
)

func ServeHttp() {
	DownloadFrontend()
	serveFrontend := http.FileServer(http.Dir("frontend"))
	fileMatcher := regexp.MustCompile(`\.[a-zA-Z]*$`)
	api.HandleApi()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !fileMatcher.MatchString(r.URL.Path) {
			http.ServeFile(w, r, "frontend/index.html")
		} else {
			serveFrontend.ServeHTTP(w, r)
		}
	})
}

func InitFrontend(ip string, port int) {
	pretty.Set(pretty.BgHiCyan, pretty.FgBlack).Print("================")
	pretty.Unset()
	pretty.Set(pretty.BgHiMagenta, pretty.FgBlack, pretty.Italic).Print(" LedFx-Frontend by Blade ")
	pretty.Unset()
	pretty.Set(pretty.BgHiCyan, pretty.FgBlack).Print("================")
	pretty.Unset()
	fmt.Print("\n")
	pretty.Set(pretty.BgBlack).Print("                                                         ")
	pretty.Unset()
	fmt.Print("\n")
	pretty.Set(pretty.BgBlack).Print("    ")
	pretty.Unset()
	switch runtime.GOOS {
	case "darwin":
		pretty.Set(pretty.BgBlack, pretty.FgWhite).Print("[CMD]+LMB: ")
	default:
		pretty.Set(pretty.BgBlack, pretty.FgWhite).Print("[CTRL]+Click: ")
	}
	pretty.Unset()
	pretty.Set(pretty.BgBlack, pretty.FgHiBlue, pretty.Bold, pretty.Underline).Print("http://localhost:8080/#/?newCore=1")
	pretty.Unset()
	switch runtime.GOOS {
	case "darwin":
		pretty.Set(pretty.BgBlack).Print("        ")
	default:
		pretty.Set(pretty.BgBlack).Print("     ")
	}
	pretty.Unset()
	fmt.Print("\n")
	pretty.Set(pretty.BgBlack).Print("                                                         ")
	pretty.Unset()
	fmt.Print("\n")
	pretty.Set(pretty.BgHiCyan, pretty.FgBlack).Print("=========================================================")
	pretty.Unset()
	fmt.Print("\n")

	go func() {
		mux := http.DefaultServeMux
		err := bridgeapi.NewServer(func(buf audio.Buffer) {
			// No callback for now
		}, mux)
		if err != nil {
			log.Logger.Fatal(err)
		}

		if err = http.ListenAndServe(fmt.Sprintf("%s:%d", ip, port), cors.AllowAll().Handler(http.DefaultServeMux)); err != nil {
			log.Logger.Fatal(err)
		}
	}()
}
