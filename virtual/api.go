package virtual

import (
	"encoding/json"
	"ledfx/config"
	"ledfx/util"
	"net/http"
)

type connectJSON struct {
	EffectID  string `json:"effect_id"`
	VirtualID string `json:"virtual_id"`
	DeviceID  string `json:"device_id"`
}

func NewAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/virtuals/schema", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get schema
			schemaBytes, err := JsonSchema()
			if util.InternalError("Virtuals API", err, writer) {
				return
			}
			writer.Write(schemaBytes)
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/virtuals/connect", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			connects := map[string]interface{}{
				"effects": connectionsEffect,
				"devices": connectionsDevice,
			}
			b, err := json.Marshal(connects)
			if util.InternalError("Virtuals API", err, writer) {
				return
			}
			writer.Write(b)
		case http.MethodPost:
			data := connectJSON{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if util.BadRequest("Virtuals API", err, writer) {
				return
			}
			if data.DeviceID != "" {
				err = ConnectDevice(data.DeviceID, data.VirtualID)
				if util.BadRequest("Virtuals API", err, writer) {
					return
				}
			}
			if data.EffectID != "" {
				err = ConnectEffect(data.EffectID, data.VirtualID)
				if util.BadRequest("Virtuals API", err, writer) {
					return
				}
			}
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/virtuals/disconnect", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusNotImplemented)
			return
		}

		data := connectJSON{}
		err := json.NewDecoder(request.Body).Decode(&data)
		if util.BadRequest("Virtuals API", err, writer) {
			return
		}
		if data.DeviceID != "" {
			err = DisconnectDevice(data.DeviceID, data.VirtualID)
			if util.BadRequest("Virtuals API", err, writer) {
				return
			}
		}
		if data.EffectID != "" {
			err = DisconnectEffect(data.EffectID, data.VirtualID)
			if util.BadRequest("Virtuals API", err, writer) {
				return
			}
		}
	})

	// handle virtual state
	mux.HandleFunc("/api/virtuals/state", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get state of all virtuals
			b, err := json.Marshal(GetStates())
			if util.InternalError("Virtuals API", err, writer) {
				return
			}
			writer.Write(b)
		case http.MethodPost:
			// Set state of all virtuals
			states := map[string]bool{}
			err := json.NewDecoder(request.Body).Decode(&states)
			if util.BadRequest("Virtuals API", err, writer) {
				return
			}
			err = SetStates(states)
			if util.InternalError("Virtuals API", err, writer) {
				return
			}
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/virtuals", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get virtuals from config
			b, err := json.Marshal(config.GetVirtuals())
			if util.InternalError("Virtuals API", err, writer) {
				return
			}
			writer.Write(b)

		case http.MethodPost:
			// Create a virtual
			data := config.VirtualEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if util.BadRequest("Virtuals API", err, writer) {
				return
			}
			_, id, err := New(data.ID, data.Config)
			if util.InternalError("Virtuals API", err, writer) {
				return
			}
			c, err := config.GetVirtual(id)
			if util.InternalError("Virtuals API", err, writer) {
				return
			}
			b, err := json.Marshal(c)
			if util.InternalError("Virtuals API", err, writer) {
				return
			}
			writer.Write(b)
			return
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})
}
