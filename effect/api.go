package effect

import (
	"encoding/json"
	"ledfx/config"
	"ledfx/util"
	"net/http"
)

func NewAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/effects/schema", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get schema
			schemaBytes, err := JsonSchema()
			if util.InternalError("Effects API", err, writer) {
				return
			}
			writer.Write(schemaBytes)
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/effects", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get effects
			b, err := json.Marshal(config.GetEffects())
			if util.InternalError("Effects API", err, writer) {
				return
			}
			writer.Write(b)

		case http.MethodPut:
			// Update an effect's config
			data := config.EffectEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if util.BadRequest("Effects API", err, writer) {
				return
			}
			effect, err := Get(data.ID)
			if util.BadRequest("Effects API", err, writer) {
				return
			}
			err = effect.UpdateBaseConfig(data.BaseConfig)
			if util.InternalError("Effects API", err, writer) {
				return
			}
			c, _ := config.GetEffect(data.ID)
			b, err := json.Marshal(c)
			if util.InternalError("Effects API", err, writer) {
				return
			}
			writer.Write(b)
			return

		case http.MethodPost:
			// Create an effect
			data := config.EffectEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if util.BadRequest("Effects API", err, writer) {
				return
			}
			_, id, err := New(data.ID, data.Type, 100, data.BaseConfig)
			if util.InternalError("Effects API", err, writer) {
				return
			}
			c, err := config.GetEffect(id)
			if util.InternalError("Effects API", err, writer) {
				return
			}
			b, err := json.Marshal(c)
			if util.InternalError("Effects API", err, writer) {
				return
			}
			writer.Write(b)
			return

		case http.MethodDelete:
			// Delete an effect
			data := config.EffectEntry{}
			keys, ok := request.URL.Query()["id"]
			if !ok || len(keys) == 0 {
				err := json.NewDecoder(request.Body).Decode(&data)
				if util.BadRequest("Effects API", err, writer) {
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

	mux.HandleFunc("/api/effects/global", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get effects global settings
			b, err := json.Marshal(globalConfig)
			if util.InternalError("Effects API", err, writer) {
				return
			}
			writer.Write(b)

		case http.MethodPut:
			// Update effects global settings
			data := make(map[string]interface{})
			err := json.NewDecoder(request.Body).Decode(&data)
			if util.BadRequest("Effects API", err, writer) {
				return
			}
			err = SetGlobalSettings(data)
			if util.InternalError("Effects API", err, writer) {
				return
			}
			c := config.GetEffectsGlobal()
			b, err := json.Marshal(c)
			if util.InternalError("Effects API", err, writer) {
				return
			}
			writer.Write(b)
			return
		}
	})
}
