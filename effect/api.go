package effect

import (
	log "ledfx/logger"
	"net/http"
)

func NewAPI(mux *http.ServeMux) {
	// TODO Handle API paths here
	mux.HandleFunc("/api/effects", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			schemaBytes, err := JsonSchema()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				log.Logger.WithField("category", "Effects API").Errorf("Error generating JSON Schema")
				return
			}
			_, _ = writer.Write(schemaBytes)
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})
}
