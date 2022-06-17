package effect

import (
	"encoding/json"
	"ledfx/config"
	log "ledfx/logger"
	"net/http"
)

func NewAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/effects/schema", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get schema
			schemaBytes, err := JsonSchema()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				log.Logger.WithField("context", "Effects API").Errorf("Error generating JSON Schema")
				return
			}
			_, _ = writer.Write(schemaBytes)
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/effects", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get effects
			b, err := json.Marshal(config.GetEffects())
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				log.Logger.WithField("context", "Effects API").Errorf("Error generating effects config")
				return
			}
			_, _ = writer.Write(b)

		case http.MethodPut:
			// Update an effect's config
			data := config.EffectEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			effect, err := Get(data.ID)
			if err != nil {
				writer.WriteHeader(http.StatusNotFound)
				writer.Write([]byte(err.Error()))
				log.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			err = effect.UpdateBaseConfig(data.BaseConfig)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
				log.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			c, _ := config.GetEffect(data.ID)
			b, err := json.Marshal(c)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				log.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			writer.Write(b)
			return

		case http.MethodPost:
			// Create an effect
			data := config.EffectEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			_, id, err := New(data.ID, data.Type, 100, data.BaseConfig)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				log.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			c, _ := config.GetEffect(id)
			b, err := json.Marshal(c)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				log.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			writer.Write(b)
			return

		case http.MethodDelete:
			// Delete an effect
			data := config.EffectEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
			}
			Destroy(data.ID)
			return
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})
}
