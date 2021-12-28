package api

import (
	"fmt"
	"io"
	"ledfx/color"
	"log"
	"net/http"
)

func InitApi(port int) error {
	// Hello world, the web server

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		_, err := io.WriteString(w, "Hello, LedFx Go!!\n")
		if err != nil {
			log.Println(err)
		}
		_, err = io.WriteString(w, "Have a good life!\n")
		if err != nil {
			log.Println(err)
		}
	}

	c := "#FF55FF"
	log.Println(color.NewColor(c))

	http.HandleFunc("/hello", helloHandler)
	log.Println("Listing for requests at http://localhost:8000/hello")
	return http.ListenAndServe(fmt.Sprint(":", port), nil)
}
