package api

import (
	"encoding/json"
	"ledfx/config"
	"net/http"
)

func HandleApi() {
	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(config.OldConfig)
	})
	http.HandleFunc("/api/goconfig", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(config.GlobalConfig)
	})
}
