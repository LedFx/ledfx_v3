package utils

import (
	"fmt"
	"ledfx/api"
	"ledfx/logger"
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

func InitFrontend() {
	fmt.Println("========================================================")
	fmt.Println("                LedFx-Frontend by Blade")
	fmt.Println("    [CTRL]+Click: http://localhost:8080/#/?newCore=1")
	fmt.Println("========================================================")
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			logger.Logger.Fatal(err)
		}
	}()
}
