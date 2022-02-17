package utils

import (
	"fmt"
	"github.com/rs/cors"
	"ledfx/api"
	"ledfx/audio"
	"ledfx/bridgeapi"
	log "ledfx/logger"
	"net/http"
	"regexp"
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
	fmt.Println("========================================================")
	fmt.Println("                LedFx-Frontend by Blade")
	fmt.Println("    [CTRL]+Click: http://localhost:8080/#/?newCore=1")
	fmt.Println("========================================================")
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
