package api

import (
	"encoding/json"
	"ledfx/config"
	"net/http"
)

func HandleApi() {
	http.HandleFunc("/api/oldconfig", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(config.OldConfig)
	})
	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(config.GlobalConfig)
	})
	http.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(config.GlobalConfig.Devices)
	})
	http.HandleFunc("/api/virtuals", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(config.GlobalConfig.Virtuals)
	})
}
