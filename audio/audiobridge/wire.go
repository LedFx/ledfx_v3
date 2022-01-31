package audiobridge

import (
	"fmt"
	"github.com/hajimehoshi/oto"
	"ledfx/integrations/airplay2"
	log "ledfx/logger"
)

func (br *Bridge) wireAirPlayOutput(client *airplay2.Client) error {
	switch br.inputType {
	case -1:
		return fmt.Errorf("input source has not been defined")
	case inputTypeAirPlayServer:
		br.airplay.server.AddClient(client)
	case inputTypeLocal:
		br.local.loopback.AddOutput(client)
		log.Logger.Fatalf("implement me!")
	}
	return fmt.Errorf("unrecognized input type")
}

func (br *Bridge) wireLocalOutput(player *oto.Player) error {
	switch br.inputType {
	case -1:
		return fmt.Errorf("input source has not been defined")
	case inputTypeAirPlayServer:
		br.airplay.server.AddOutput(player)
	case inputTypeLocal:
		br.local.loopback.AddOutput(player)
		log.Logger.Fatalf("implement me!")
	}
	return fmt.Errorf("unrecognized input type")
}
