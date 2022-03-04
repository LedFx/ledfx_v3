package bridgeapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ledfx/config"
	"ledfx/virtual"
	"net/http"
	"path/filepath"
	"strings"
)

type SetVirtualEffectReq struct {
	Active bool               `json:"active"`
	Type   string             `json:"type"`
	Config VirtualColorConfig `json:"config"`
}
type VirtualColorConfig struct {
	Color string `json:"color"`
}

func (s *Server) HandleVirtuals(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errToJson(err))
		return
	}

	switch filepath.Base(r.URL.Path) {
	case "effects":
		split := strings.Split(r.URL.Path, "/")
		if len(split) != 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errToJson(err))
			return
		}

		effectReq := SetVirtualEffectReq{}
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
	}
}

func (s *Server) setVirtualEffect(effectReq *SetVirtualEffectReq, virtualID string) error {
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
