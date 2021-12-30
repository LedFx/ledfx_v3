package api

import (
	"encoding/json"
	"ledfx/config"
	"ledfx/logger"
	"ledfx/virtual"
	"net/http"
	"strings"
)

func SetHeader(w http.ResponseWriter) {
	headers := w.Header()
	headers.Add("Access-Control-Allow-Origin", "*")
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")
	headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
	headers.Add("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
}

// func SetHeader(w http.ResponseWriter) {
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
// }

type Resp struct {
	Active bool
}

func HandleApi() {
	http.HandleFunc("/api/oldconfig", func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)
		json.NewEncoder(w).Encode(config.OldConfig)
	})
	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)
		json.NewEncoder(w).Encode(config.GlobalConfig)
	})
	http.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)
		json.NewEncoder(w).Encode(config.GlobalConfig)
		// json.NewEncoder(w).Encode(config.GlobalConfig.Devices)
	})
	http.HandleFunc("/api/virtuals", func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)
		json.NewEncoder(w).Encode(config.GlobalConfig)
		// json.NewEncoder(w).Encode(config.GlobalConfig.Virtuals)
	})
	http.HandleFunc("/api/virtuals/", func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)
		logger.Logger.Debug(r.Method)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			var p Resp

			path := strings.TrimPrefix(r.URL.Path, "/virtuals/")
			virtualid := strings.Split(path, "/api/virtuals/")[1]

			err := json.NewDecoder(r.Body).Decode(&p)
			if err != nil {
				logger.Logger.Warn(err)
				// http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			virtual.PlayVirtual(virtualid, p.Active)
			json.NewEncoder(w).Encode(config.GlobalConfig.Virtuals)

		}
	})
	http.HandleFunc("/api/schema", func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)
		json.NewEncoder(w).Encode(config.GlobalConfig)
		// json.NewEncoder(w).Encode(config.Schema)
	})
}
