package config

import (
	"encoding/json"
	"net/http"

	"github.com/LedFx/ledfx/pkg/util"
)

func NewAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/settings/schema", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get schema
			schemaBytes, err := CoreJsonSchema()
			if util.InternalError("Settings API", err, writer) {
				return
			}
			writer.Write(schemaBytes)
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
			if util.InternalError("Settings API", err, writer) {
				return
			}
			writer.Write(b)

		case http.MethodPut:
			// Update settings
			settings := make(map[string]interface{})
			err := json.NewDecoder(request.Body).Decode(&settings)
			if util.BadRequest("Settings API", err, writer) {
				return
			}
			err = SetSettings(settings)
			if util.BadRequest("Settings API", err, writer) {
				return
			}
			b, err := json.Marshal(store.Settings)
			if util.InternalError("Settings API", err, writer) {
				return
			}
			writer.Write(b)
			return

		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})
}
