package color

import (
	"encoding/json"
	"ledfx/util"
	"net/http"
)

func NewAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/color", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get colors and gradients
			res := make(map[string]interface{})
			res["colors"] = LedFxColors
			res["palettes"] = LedFxPalettes
			resBytes, err := json.MarshalIndent(res, "", "\t")
			if util.InternalError("Color API", err, writer) {
				return
			}
			writer.Write(resBytes)
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})
}
