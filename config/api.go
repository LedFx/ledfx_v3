package config

import (
	"encoding/json"
	log "ledfx/logger"
	"net/http"
)

func NewAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/settings/schema", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get schema
			schemaBytes, err := CoreJsonSchema()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				log.Logger.WithField("context", "Settings API").Error("Error generating JSON Schema:", err)
				return
			}
			_, _ = writer.Write(schemaBytes)
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/settings", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get settings
			// NOTE this gets the settings including any command line flags
			// use store.settings to get the settings saved in config
			// i think the active settings make most sense here, as long as the frontend sends incremental updates
			// if the frontend sends all the settings back, then settings modified by command line flags will be saved to config
			b, err := json.Marshal(GetSettings())
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				log.Logger.WithField("context", "Settings API").Errorf("Error generating settings config")
				return
			}
			_, _ = writer.Write(b)

		case http.MethodPut:
			// Update settings
			settings := make(map[string]interface{})
			err := json.NewDecoder(request.Body).Decode(&settings)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			err = SetSettings(settings)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			b, err := json.Marshal(store.Settings)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				log.Logger.WithField("context", "Settings API").Error(err)
				return
			}
			writer.Write(b)
			return

		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})
}
