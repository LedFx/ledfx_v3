package utils

import (
	"net/http"
)

func SetupRoutes() {
	GetFrontend()
	ServeFrontend()
	// map our `/ws` endpoint to the `serveWs` function
	http.HandleFunc("/ws", ServeWs)
}
