package bridgeapi

import (
	"net/http"
)

type VirtualData struct {
	LastColor string `json:"color"`
}

type VirtualEffect struct {
	Active bool                `json:"active"`
	Type   string              `json:"type"`
	Config VirtualEffectConfig `json:"config"`
}
type VirtualEffectConfig struct {
	Color string `json:"color"`
}

func (s *Server) setVirtualEffect(effectReq *VirtualEffect, virtualID string) error { return nil }

func (s *Server) HandleVirtuals(w http.ResponseWriter, r *http.Request) {}

// COMMENTED THIS OUT; MOST OF ITS METHODS ARE REDUNDANT

// func (s *Server) HandleVirtuals(w http.ResponseWriter, r *http.Request) {
// 	api.SetHeader(w)

// 	if r.Method == http.MethodOptions {
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}

// 	var p VirtualEffect

// 	// Ex. r.URL.Path == "/api/virtuals/yz-quad-1/effects" (or with trailing slash)

// 	// Ex. path == "yz-quad-1/effects"
// 	path := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/virtuals/"), "/")

// 	// Ex. split == ["yz-quad-1", "effects"]
// 	split := strings.Split(path, "/")

// 	// Check if bounds are correct
// 	if len(split) != 2 {
// 		http.Error(w, "invalid path format", http.StatusNotFound)
// 		return
// 	}

// 	// Ex. virtualID == split[0] ("yz-quad-1")
// 	virtualID := split[0]
// 	// Ex. category == split[1] ("effects")
// 	category := split[1]

// 	defer r.Body.Close()
// 	err := json.NewDecoder(r.Body).Decode(&p)
// 	if err != nil {
// 		logger.Logger.Warn(err)
// 		return
// 	}

// 	switch r.Method {
// 	case http.MethodDelete:
// 		if err := virtual.StopVirtual(virtualID); err != nil {
// 			logger.Logger.WithField("category", fmt.Sprintf("HTTP/DELETE: %s", r.URL.Path)).Warnf("Error stopping virtual with ID %q: %v", virtualID, err)
// 		}
// 	case http.MethodPost, http.MethodPut:
// 		for i, d := range config.GlobalConfig.Virtuals {
// 			if d.Id == virtualID {
// 				config.GlobalConfig.Virtuals[i].Effect.Type = p.Type
// 				config.GlobalConfig.Virtuals[i].Effect.Name = p.Type // ToDo @carterpeel: Take Name from refactored schema.go line 46

// 			}
// 		}
// 		if p.Type == "audioRandom" {
// 			p.Active = true
// 			p.Config.Color = color.RandomColor()
// 			s.virtuals.LastColor = color.RandomColor()
// 		}

// 		switch category {
// 		case "effects":
// 			s.virtuals.LastColor = p.Config.Color
// 			for i, d := range config.GlobalConfig.Virtuals {
// 				if d.Id == virtualID {
// 					config.GlobalConfig.Virtuals[i].Effect.Config.Color = s.virtuals.LastColor
// 				}
// 			}
// 			if err := virtual.PlayVirtual(virtualID, true, s.virtuals.LastColor, p.Type); err != nil {
// 				logger.Logger.WithField("category", fmt.Sprintf("HTTP/%s: %s", r.Method, r.URL.Path)).Warnf("Error playing virtual effect with ID %q: %v", virtualID, err)
// 			}
// 		case "presets":
// 			logger.Logger.WithField("category", fmt.Sprintf("HTTP/%s: %s", r.Method, r.URL.Path)).Debug("No Presets yet ;)")
// 		default:
// 			if s.virtuals.LastColor == "" {
// 				s.virtuals.LastColor = "#000fff"
// 			}
// 			err := virtual.PlayVirtual(virtualID, p.Active, s.virtuals.LastColor, p.Type)
// 			if err != nil {
// 				logger.Logger.WithField("category", fmt.Sprintf("HTTP/%s: %s", r.Method, r.URL.Path)).Warnf("Error playing virtual effect with ID %q: %v", virtualID, err)
// 			}
// 		}

// 		if err := json.NewEncoder(w).Encode(config.GlobalConfig.Virtuals); err != nil {
// 			logger.Logger.WithField("category", fmt.Sprintf("HTTP/%s: %s", r.Method, r.URL.Path)).Warnf("Error encoding 'GlobalConfig.Virtuals': %v", err)
// 		}
// 	}

// 	/*defer r.Body.Close()
// 	bodyBytes, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		w.Write(errToJson(err))
// 		return
// 	}

// 	switch filepath.Base(r.URL.Path) {
// 	case "effects":
// 		split := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
// 		if len(split) != 3 {
// 			w.WriteHeader(http.StatusNotFound)
// 			return
// 		}

// 		effectReq := VirtualEffect{}
// 		if err := json.Unmarshal(bodyBytes, &effectReq); err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write(errToJson(err))
// 			return
// 		}
// 		if err := s.setVirtualEffect(&effectReq, split[1]); err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write(errToJson(err))
// 			return
// 		}
// 	default:
// 		w.WriteHeader(http.StatusNotFound)
// 	}*/
// }
