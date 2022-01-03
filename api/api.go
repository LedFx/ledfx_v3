package api

import (
	_ "embed"
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

type RespConfig struct {
	Color string `json:"color"`
}
type Resp struct {
	Active bool       `json:"active"`
	Config RespConfig `json:"config"`
	Type   string     `json:"type"`
}

var LastColor string

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
		// TODO: See comment for Virtuals
		// json.NewEncoder(w).Encode(config.GlobalConfig.Devices)
	})

	http.HandleFunc("/api/virtuals", func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)
		// TODO:
		// this is too much, we only need Virtuals
		json.NewEncoder(w).Encode(config.GlobalConfig)

		// this is too less, we need the key also: {"virtuals": ...}
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
			var category string
			var virtualid string
			path := strings.TrimPrefix(r.URL.Path, "/virtuals/")
			virtualid = strings.Split(path, "/api/virtuals/")[1]
			pathNodes := strings.Split(virtualid, "/")
			if len(pathNodes) > 1 {
				category = string(pathNodes[1])
				virtualid = string(pathNodes[0])
			}

			err := json.NewDecoder(r.Body).Decode(&p)
			if err != nil {
				logger.Logger.Warn(err)
				// http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// logger.Logger.Debug(p)
			if category == "effects" {
				// logger.Logger.Debug(p.Config.Color)
				LastColor = p.Config.Color
				virtual.PlayVirtual(virtualid, true, LastColor)
			} else if category == "presets" {
				logger.Logger.Debug("No Presets yet ;)")
				// virtual.PlayVirtual(virtualid, p.Active, "#fff000")
				// utils.SendWs(utils.Ws, "warning", "No Presets yet ;)")
			} else {
				if LastColor == "" {
					LastColor = "#000fff"
				}
				virtual.PlayVirtual(virtualid, p.Active, LastColor)
			}

			json.NewEncoder(w).Encode(config.GlobalConfig.Virtuals)

		}
	})
	HandleSchema()
	HandleColors()

	// http.HandleFunc("/api/schema", func(w http.ResponseWriter, r *http.Request) {
	// 	SetHeader(w)
	// 	json.NewEncoder(w).Encode(config.GlobalConfig)
	// 	// json.NewEncoder(w).Encode(config.Schema)
	// })

}
