package device

import (
	"encoding/json"
	"ledfx/config"
	"ledfx/logger"
	"net/http"
)

func NewAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/devices/schema", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get schema
			schemaBytes, err := JsonSchema()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				logger.Logger.WithField("context", "Devices API").Errorf("Error generating JSON Schema")
				return
			}
			_, _ = writer.Write(schemaBytes)
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/devices", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get devices
			b, err := json.Marshal(config.GetDevices())
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Devices API").Errorf("Error generating devices config")
				return
			}
			_, _ = writer.Write(b)

		case http.MethodPost:
			// Create a device
			data := config.DeviceEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
				return
			}
			_, id, err := New(data.ID, data.Type, data.BaseConfig, data.ImplConfig)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Devices API").Error(err)
				return
			}
			c, err := config.GetDevice(id)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Devices API").Error(err)
				return
			}
			b, err := json.Marshal(c)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Devices API").Error(err)
				return
			}
			writer.Write(b)
			return
		}

	})
}
