package api

import (
	"fmt"
	"io"
	"ledfx/logger"
	"net/http"
)

func InitApi(port int) error {
	// Hello world, the web server

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(w, "Hello, LedFx Go!!\n")
		if err != nil {
			logger.Logger.Panic(err)
		}
		_, err = io.WriteString(w, "Have a good life!\n")
		if err != nil {
			logger.Logger.Panic(err)
		}
	}

	http.HandleFunc("/hello", helloHandler)
	// TODO: change to config.Host
	logger.Logger.Debug("Listing for requests at http://localhost:8000/hello")
	return http.ListenAndServe(fmt.Sprint(":", port), nil)
}
