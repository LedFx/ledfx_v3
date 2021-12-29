package utils

import (
	"net/http"
)

func SetupRoutes() {
	DownloadFrontend()
	ServeFrontend()
	// map our `/ws` endpoint to the `serveWs` function
	http.HandleFunc("/ws", ServeWs)
}
