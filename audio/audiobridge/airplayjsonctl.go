package audiobridge

import (
	"encoding/json"
	"fmt"
	"ledfx/integrations/airplay2"
)

type AirPlayAction string

const (
	AirPlayActionStopServer AirPlayAction = "stop"
	AirPlayActionGetClients               = "clients"
)

type AirPlayCTLJSON struct {
	Action AirPlayAction `json:"action"`
}

func (apctl AirPlayCTLJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&apctl)
}

type ClientList struct {
	Clients []*airplay2.Client `json:"clients"`
}

func (cl *ClientList) AsJSON() ([]byte, error) {
	return json.Marshal(cl)
}

// AirPlay takes a marshalled AirPlayCTLJSON
//
// If AirPlayCTLJSON.Action == AirPlayActionStopServer, the server will stop.
//
// If AirPlayCTLJSON.Action == AirPlayActionGetClients, the first return value will be non-nil.
func (j *JsonCTL) AirPlay(jsonData []byte) (clients *ClientList, err error) {
	conf := AirPlayCTLJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	switch conf.Action {
	case AirPlayActionStopServer:
		return nil, j.w.br.Controller().AirPlay().StopServer()
	case AirPlayActionGetClients:
		return &ClientList{
			Clients: j.w.br.Controller().AirPlay().Clients(),
		}, nil
	}

	return nil, fmt.Errorf("unknown action '%d'", conf.Action)
}
