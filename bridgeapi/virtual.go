package bridgeapi

import (
	"encoding/json"
	"fmt"
	"ledfx/api"
	"ledfx/config"
	"ledfx/logger"
	"ledfx/virtual"
	"net/http"
	"strings"
)

type VirtualData struct {
	LastColor string
}

type VirtualEffect struct {
	Active bool                `json:"active"`
	Type   string              `json:"type"`
	Config VirtualEffectConfig `json:"config"`
}
type VirtualEffectConfig struct {
	Color string `json:"color"`
}

func (s *Server) HandleVirtuals(w http.ResponseWriter, r *http.Request) {
	api.SetHeader(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var p VirtualEffect
	var category string
	var virtualID string

	path := strings.TrimPrefix(r.URL.Path, "/virtuals/")
	if spl := strings.Split(path, "/api/virtuals"); len(spl) > 1 {
		virtualID = spl[1]
	}
	pathNodes := strings.Split(virtualID, "/")

	if len(pathNodes) > 1 {
		category = pathNodes[1]
		virtualID = pathNodes[0]
	}

	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		logger.Logger.Warn(err)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		if err := virtual.StopVirtual(virtualID); err != nil {
			logger.Logger.WithField("category", fmt.Sprintf("HTTP/DELETE: %s", r.URL.Path)).Warnf("Error stopping virtual with ID %q: %v", virtualID, err)
		}
	case http.MethodPost, http.MethodPut:
		switch category {
		case "effects":
			s.virtuals.LastColor = p.Config.Color
			if err := virtual.PlayVirtual(virtualID, true, s.virtuals.LastColor); err != nil {
				logger.Logger.WithField("category", fmt.Sprintf("HTTP/%s: %s", r.Method, r.URL.Path)).Warnf("Error playing virtual effect with ID %q: %v", virtualID, err)
			}
		case "presets":
			logger.Logger.WithField("category", fmt.Sprintf("HTTP/%s: %s", r.Method, r.URL.Path)).Debug("No Presets yet ;)")
		default:
			if s.virtuals.LastColor == "" {
				s.virtuals.LastColor = "#000fff"
			}
			err := virtual.PlayVirtual(virtualID, p.Active, s.virtuals.LastColor)
			if err != nil {
				logger.Logger.WithField("category", fmt.Sprintf("HTTP/%s: %s", r.Method, r.URL.Path)).Warnf("Error playing virtual effect with ID %q: %v", virtualID, err)
			}
		}

		if err := json.NewEncoder(w).Encode(config.GlobalConfig.Virtuals); err != nil {
			logger.Logger.WithField("category", fmt.Sprintf("HTTP/%s: %s", r.Method, r.URL.Path)).Warnf("Error encoding 'GlobalConfig.Virtuals': %v", err)
		}
	}

	/*defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errToJson(err))
		return
	}

	switch filepath.Base(r.URL.Path) {
	case "effects":
		split := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if len(split) != 3 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		effectReq := VirtualEffect{}
		if err := json.Unmarshal(bodyBytes, &effectReq); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errToJson(err))
			return
		}
		if err := s.setVirtualEffect(&effectReq, split[1]); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errToJson(err))
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}*/
}

func (s *Server) setVirtualEffect(effectReq *VirtualEffect, virtualID string) error {
	virtualIndex := -1
	for index, virt := range config.GlobalConfig.Virtuals {
		if virt.Id == virtualID {
			virtualIndex = index
			break
		}
	}

	if virtualIndex == -1 {
		return fmt.Errorf("virtual with id %q not found", virtualID)
	}

	config.GlobalConfig.Virtuals[virtualIndex].Active = effectReq.Active
	config.GlobalConfig.Virtuals[virtualIndex].Effect.Type = effectReq.Type
	config.GlobalConfig.Virtuals[virtualIndex].Effect.Config.Color = effectReq.Config.Color
	return virtual.PlayVirtual(virtualID, effectReq.Active, effectReq.Config.Color)
}
