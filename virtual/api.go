package virtual

import (
	"encoding/json"
	"ledfx/config"
	"ledfx/logger"
	"net/http"
)

func NewAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/virtuals/schema", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get schema
			schemaBytes, err := JsonSchema()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				logger.Logger.WithField("context", "Virtuals API").Errorf("Error generating JSON Schema")
				return
			}
			_, _ = writer.Write(schemaBytes)
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/virtuals", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get virtuals
			b, err := json.Marshal(config.GetVirtuals())
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				logger.Logger.WithField("context", "Virtuals API").Errorf("Error generating effects config")
				return
			}
			_, _ = writer.Write(b)

		case http.MethodPost:
			// Create a virtual
			data := config.VirtualEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if err != nil {
				writer.Write([]byte(err.Error()))
				writer.WriteHeader(http.StatusBadRequest)
				return
			}
			_, id, err := New(data.ID, data.Config)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Virtuals API").Error(err)
				return
			}
			c, err := config.GetVirtual(id)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Virtuals API").Error(err)
				return
			}
			b, err := json.Marshal(c)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Virtuals API").Error(err)
				return
			}
			writer.Write(b)
			return
		}

	})
}
