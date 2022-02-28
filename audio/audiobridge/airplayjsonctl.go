package audiobridge

import (
	"encoding/json"
	"fmt"
	"ledfx/integrations/airplay2"
)

type AirPlayAction string

const (
	AirPlayActionStopServer AirPlayAction = "stop"
)

type AirPlayCTLJSON struct {
	Action AirPlayAction `json:"action"`
}

func (apctls AirPlayCTLJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&apctls)
}

type ClientList struct {
	Clients []*airplay2.Client `json:"clients"`
}

func (cl *ClientList) AsJSON() ([]byte, error) {
	return json.Marshal(cl)
}

// AirPlaySet takes a marshalled AirPlayCTLJSON
func (j *JsonCTL) AirPlaySet(jsonData []byte) (err error) {
	conf := AirPlayCTLJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	switch conf.Action {
	case AirPlayActionStopServer:
		return j.w.br.Controller().AirPlay().StopServer()
	}

	return fmt.Errorf("unknown action '%s'", conf.Action)
}

func (j *JsonCTL) AirPlayGetClients() (resultJson []byte, err error) {
	cList := &ClientList{
		Clients: j.w.br.Controller().AirPlay().Clients(),
	}
	return cList.AsJSON()
}
