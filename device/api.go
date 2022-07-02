package device

import (
	"encoding/json"
	"ledfx/config"
	"ledfx/util"
	"net/http"
)

func NewAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/devices/schema", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get schema
			schemaBytes, err := JsonSchema()
			if util.InternalError("Device API", err, writer) {
				return
			}
			writer.Write(schemaBytes)
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/devices/state", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get schema
			s, err := json.Marshal(GetStates())
			if util.InternalError("Device API", err, writer) {
				return
			}
			writer.Write(s)
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/devices", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get devices
			b, err := json.Marshal(config.GetDevices())
			if util.InternalError("Device API", err, writer) {
				return
			}
			writer.Write(b)

		case http.MethodPost:
			// Create a device
			data := config.DeviceEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if util.BadRequest("Device API", err, writer) {
				return
			}
			_, id, err := New(data.ID, data.Type, data.BaseConfig, data.ImplConfig)
			if util.InternalError("Device API", err, writer) {
				return
			}
			c, err := config.GetDevice(id)
			if util.InternalError("Device API", err, writer) {
				return
			}
			b, err := json.Marshal(c)
			if util.InternalError("Device API", err, writer) {
				return
			}
			writer.Write(b)
			return
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})
}
