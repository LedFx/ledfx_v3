package controller

import (
	"encoding/json"
	"net/http"

	"github.com/LedFx/ledfx/pkg/config"
	"github.com/LedFx/ledfx/pkg/util"
)

type connectJSON struct {
	EffectID     string `json:"effect_id"`
	ControllerID string `json:"controller_id"`
	DeviceID     string `json:"device_id"`
}

func NewAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/controllers/schema", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get schema
			schemaBytes, err := JsonSchema()
			if util.InternalError("Controllers API", err, writer) {
				return
			}
			writer.Write(schemaBytes)
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/controllers/connect", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			connects := map[string]interface{}{
				"effects": connectionsEffect,
				"devices": connectionsDevice,
			}
			b, err := json.Marshal(connects)
			if util.InternalError("Controllers API", err, writer) {
				return
			}
			writer.Write(b)
		case http.MethodPost:
			data := connectJSON{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if util.BadRequest("Controllers API", err, writer) {
				return
			}
			if data.DeviceID != "" {
				err = ConnectDevice(data.DeviceID, data.ControllerID)
				if util.BadRequest("Controllers API", err, writer) {
					return
				}
			}
			if data.EffectID != "" {
				err = ConnectEffect(data.EffectID, data.ControllerID)
				if util.BadRequest("Controllers API", err, writer) {
					return
				}
			}
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/controllers/disconnect", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusNotImplemented)
			return
		}

		data := connectJSON{}
		err := json.NewDecoder(request.Body).Decode(&data)
		if util.BadRequest("Controllers API", err, writer) {
			return
		}
		if data.DeviceID != "" {
			err = DisconnectDevice(data.DeviceID, data.ControllerID)
			if util.BadRequest("Controllers API", err, writer) {
				return
			}
		}
		if data.EffectID != "" {
			err = DisconnectEffect(data.EffectID, data.ControllerID)
			if util.BadRequest("Controllers API", err, writer) {
				return
			}
		}
	})

	// handle controller state
	mux.HandleFunc("/api/controllers/state", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get state of all controllers
			b, err := json.Marshal(GetStates())
			if util.InternalError("Controllers API", err, writer) {
				return
			}
			writer.Write(b)
		case http.MethodPost:
			// Set state of all controllers
			states := map[string]bool{}
			err := json.NewDecoder(request.Body).Decode(&states)
			if util.BadRequest("Controllers API", err, writer) {
				return
			}
			err = SetStates(states)
			if util.InternalError("Controllers API", err, writer) {
				return
			}
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/controllers", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get controllers from config
			b, err := json.Marshal(config.GetControllers())
			if util.InternalError("Controllers API", err, writer) {
				return
			}
			writer.Write(b)

		case http.MethodPost:
			// Create a controller
			data := config.ControllerEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if util.BadRequest("Controllers API", err, writer) {
				return
			}
			_, id, err := New(data.ID, data.Config)
			if util.InternalError("Controllers API", err, writer) {
				return
			}
			c, err := config.GetController(id)
			if util.InternalError("Controllers API", err, writer) {
				return
			}
			b, err := json.Marshal(c)
			if util.InternalError("Controllers API", err, writer) {
				return
			}
			writer.Write(b)
			return
		case http.MethodDelete:
			// Delete a controller
			data := config.ControllerEntry{}
			keys, ok := request.URL.Query()["id"]
			if !ok || len(keys) == 0 {
				err := json.NewDecoder(request.Body).Decode(&data)
				if util.BadRequest("Controllers API", err, writer) {
					return
				}
			} else {
				data.ID = keys[0]
			}
			Destroy(data.ID)
			return
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})
}
