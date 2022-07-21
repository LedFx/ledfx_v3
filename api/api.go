package api

import (
	_ "embed"
	"encoding/json"
	"ledfx/audio"
	"ledfx/logger"
	"net/http"
)

func SetHeader(w http.ResponseWriter) {
	headers := w.Header()
	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")
	headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
	headers.Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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
	http.HandleFunc("/api/audio", func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w)
		audioDevices, err := audio.GetAudioDevices()
		if err != nil {
			logger.Logger.Warn(err)
		}
		err = json.NewEncoder(w).Encode(audioDevices)
		if err != nil {
			logger.Logger.Warn(err)
		}
	})
	// http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
	// 	SetHeader(w)
	// 	err := json.NewEncoder(w).Encode(config.GlobalConfig)
	// 	if err != nil {
	// 		logger.Logger.Warn(err)
	// 	}
	// })

	// http.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
	// 	SetHeader(w)
	// 	err := json.NewEncoder(w).Encode(config.GlobalConfig)
	// 	if err != nil {
	// 		logger.Logger.Warn(err)
	// 	}
	// 	// TODO: See comment for Controllers
	// 	// json.NewEncoder(w).Encode(config.GlobalConfig.Devices)
	// })

	// http.HandleFunc("/api/controllers", func(w http.ResponseWriter, r *http.Request) {
	// 	SetHeader(w)
	// 	// TODO:
	// 	// this is too much, we only need Controllers
	// 	err := json.NewEncoder(w).Encode(config.GlobalConfig)
	// 	if err != nil {
	// 		logger.Logger.Warn(err)
	// 	}

	// 	// this is too less, we need the key also: {"controllers": ...}
	// 	// json.NewEncoder(w).Encode(config.GlobalConfig.Controllers)
	// })
	/*	http.HandleFunc("/api/controllers/", func(w http.ResponseWriter, r *http.Request) {
			if LastColor == "" {
				LastColor = "#ff0000"
			}
			SetHeader(w)
			logger.Logger.Debug(r.Method)
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			} else {

				var p Resp
				var category string
				var controllerid string
				path := strings.TrimPrefix(r.URL.Path, "/controllers/")
				controllerid = strings.Split(path, "/api/controllers/")[1]
				pathNodes := strings.Split(controllerid, "/")
				if len(pathNodes) > 1 {
					category = string(pathNodes[1])
					controllerid = string(pathNodes[0])
				}

				err := json.NewDecoder(r.Body).Decode(&p)
				if err != nil {
					logger.Logger.Warn(err)
					return
				}

				if r.Method == "DELETE" {
					err := controller.StopController(controllerid)
					if err != nil {
						logger.Logger.Warn(err)
					}
					return
				}
				if r.Method == "POST" || r.Method == "PUT" {
					err := json.NewDecoder(r.Body).Decode(&p)
					if err != nil {
						logger.Logger.Warn(err)
						// http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
				}

				if category == "effects" {
					LastColor = p.Config.Color
					err := controller.PlayController(controllerid, true, LastColor)
					if err != nil {
						logger.Logger.Warn(err)
					}
				} else if category == "presets" {
					logger.Logger.Debug("No Presets yet ;)")
				} else {
					if LastColor == "" {
						LastColor = "#000fff"
					}
					err := controller.PlayController(controllerid, p.Active, LastColor)
					if err != nil {
						logger.Logger.Warn(err)
					}
				}

				err = json.NewEncoder(w).Encode(config.GlobalConfig.Controllers)
				if err != nil {
					logger.Logger.Warn(err)
				}

			}
		})
	*/
	HandleSchema()
	HandleColors()
}
