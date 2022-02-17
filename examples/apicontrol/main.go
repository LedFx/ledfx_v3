package main

import (
	"flag"
	"ledfx/audio"
	"log"
)

func init() {
	flag.IntVar(&port, "port", 8080, "The port the API server will listen on")
	flag.StringVar(&ip, "ip", "0.0.0.0", "The IP address the API server will listen on")
}

var (
	port int
	ip   string
)

func main() {
	server, err := NewServer(func(buf audio.Buffer) {
		// No callback for now...
	})
	if err != nil {
		log.Panicf("Error initializing API server: %v\n", err)
	}

	log.Panicf("Error starting API server: %v\n", server.Serve(ip, port))
}
