package virtual

import (
	"encoding/json"
	"errors"
	"ledfx/config"
	"ledfx/logger"
	"net/http"
)

type connectJSON struct {
	EffectID  string `json:"effect_id"`
	VirtualID string `json:"virtual_id"`
	DeviceID  string `json:"device_id"`
}

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

	mux.HandleFunc("/api/virtuals/connect", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			connects := map[string]interface{}{
				"effects": connectionsEffect,
				"devices": connectionsDevice,
			}
			b, err := json.Marshal(connects)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				logger.Logger.WithField("context", "Virtuals API").Errorf("Error generating effects config")
				return
			}
			_, _ = writer.Write(b)
		case http.MethodPost:
			data := connectJSON{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
				return
			}
			if data.VirtualID == "" {
				if err != nil {
					err = errors.New("need virtual ID")
					writer.WriteHeader(http.StatusBadRequest)
					writer.Write([]byte(err.Error()))
					return
				}
			}
			if data.DeviceID == "" || data.EffectID == "" {
				if err != nil {
					err = errors.New("need effect or device ID or both")
					writer.WriteHeader(http.StatusBadRequest)
					writer.Write([]byte(err.Error()))
					return
				}
			}
			if data.DeviceID != "" {
				err = ConnectDevice(data.DeviceID, data.VirtualID)
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					writer.Write([]byte(err.Error()))
					return
				}
			}
			if data.EffectID != "" {
				err = ConnectEffect(data.EffectID, data.VirtualID)
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					writer.Write([]byte(err.Error()))
					return
				}
			}
		default:
			writer.WriteHeader(http.StatusNotImplemented)
		}
	})

	mux.HandleFunc("/api/virtuals/disconnect", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			err := errors.New("only POST method allowed")
			writer.Write([]byte(err.Error()))
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		data := connectJSON{}
		err := json.NewDecoder(request.Body).Decode(&data)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte(err.Error()))
			return
		}
		if data.VirtualID == "" {
			if err != nil {
				err = errors.New("need virtual ID")
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
				return
			}
		}
		if data.DeviceID == "" || data.EffectID == "" {
			if err != nil {
				err = errors.New("need effect or device ID or both")
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
				return
			}
		}
		if data.DeviceID != "" {
			err = DisconnectDevice(data.DeviceID, data.VirtualID)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
				return
			}
		}
		if data.EffectID != "" {
			err = DisconnectEffect(data.EffectID, data.VirtualID)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
				return
			}
		}
	})

	// handle virtual state
	mux.HandleFunc("/api/virtuals/state", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get state of all virtuals
			b, err := json.Marshal(GetStates())
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				logger.Logger.WithField("context", "Virtuals API").Errorf("Error generating virtuals states")
				return
			}
			_, _ = writer.Write(b)
		case http.MethodPost:
			// Set state of all virtuals
			states := map[string]bool{}
			err := json.NewDecoder(request.Body).Decode(&states)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
				return
			}
			err = SetStates(states)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				logger.Logger.WithField("context", "Virtuals API").Error(err)
				return
			}
		}
	})

	mux.HandleFunc("/api/virtuals", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			// Get virtuals
			b, err := json.Marshal(config.GetVirtuals())
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				logger.Logger.WithField("context", "Virtuals API").Errorf("Error generating virtuals config")
				return
			}
			_, _ = writer.Write(b)

		case http.MethodPost:
			// Create a virtual
			data := config.VirtualEntry{}
			err := json.NewDecoder(request.Body).Decode(&data)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte(err.Error()))
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
