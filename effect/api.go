package effect

import (
	"encoding/json"
	"ledfx/config"
	"ledfx/logger"
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
				logger.Logger.WithField("context", "Effects API").Errorf("Error generating JSON Schema")
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
				logger.Logger.WithField("context", "Effects API").Errorf("Error generating effects config")
				return
			}
			_, _ = writer.Write(b)

		case http.MethodPut:
			// Update an effect's config
			data := config.EffectEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
				return
			}
			effect, err := Get(data.ID)
			if err != nil {
				writer.WriteHeader(http.StatusNotFound)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			err = effect.UpdateBaseConfig(data.BaseConfig)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			c, _ := config.GetEffect(data.ID)
			b, err := json.Marshal(c)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			writer.Write(b)
			return

		case http.MethodPost:
			// Create an effect
			data := config.EffectEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if err != nil {
				writer.Write([]byte(err.Error()))
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			_, id, err := New(data.ID, data.Type, 100, data.BaseConfig)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			c, err := config.GetEffect(id)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			b, err := json.Marshal(c)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Effects API").Error(err)
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
				writer.Write([]byte(err.Error()))
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
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				logger.Logger.WithField("context", "Effects API").Errorf("Error fetching global effects config")
				return
			}
			_, _ = writer.Write(b)

		case http.MethodPut:
			// Update effects global settings
			data := make(map[string]interface{})
			err := json.NewDecoder(request.Body).Decode(&data)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			err = SetGlobalSettings(data)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			c := config.GetEffectsGlobal()
			b, err := json.Marshal(c)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Effects API").Error(err)
				return
			}
			writer.Write(b)
			return
		}
	})
}
