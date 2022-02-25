package audiobridge

import (
	"encoding/json"
	"fmt"
	"ledfx/integrations/airplay2"
)

type AirPlaySetAction string

const (
	AirPlayActionStopServer AirPlaySetAction = "stop"
)

type AirPlayJsonCtlSet struct {
	Action AirPlaySetAction `json:"action"`
}

func (apctls AirPlayJsonCtlSet) AsJSON() ([]byte, error) {
	return json.Marshal(&apctls)
}

type ClientList struct {
	Clients []*airplay2.Client `json:"clients"`
}

func (cl *ClientList) AsJSON() ([]byte, error) {
	return json.Marshal(cl)
}

// AirPlaySet takes a marshalled AirPlayJsonCtlSet
func (j *JsonCTL) AirPlaySet(jsonData []byte) (err error) {
	conf := AirPlayJsonCtlSet{}
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
