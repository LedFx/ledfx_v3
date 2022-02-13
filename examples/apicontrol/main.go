package main

import (
	"ledfx/audio"
	"log"
)

func main() {
	server, err := NewServer(func(buf audio.Buffer) {
		// No callback for now...
	})
	if err != nil {
		log.Panicf("Error initializing API server: %v\n", err)
	}

	log.Panicf("Error starting API server: %v\n", server.Serve("127.0.0.1", 8080))
}
