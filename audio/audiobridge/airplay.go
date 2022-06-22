package audiobridge

import (
	"fmt"
	"ledfx/integrations/airplay2"
	log "ledfx/logger"
)

func (br *Bridge) StartAirPlayInput(name string, port int) error {
	if br.inputType != -1 {
		br.closeInput()
	}
	br.inputType = inputTypeAirPlayServer

	if br.airplay == nil {
		br.airplay = newAirPlayHandler()
	}

	br.airplay.server = airplay2.NewServer(airplay2.Config{
		AdvertisementName: name,
		Port:              port,
	}, br.byteWriter)

	if err := br.airplay.server.Start(); err != nil {
		return fmt.Errorf("error starting AirPlay server: %w", err)
	}
	return nil
}

func (br *Bridge) AddAirPlayOutput(searchKey string, searchType AirPlaySearchType) error {
	if br.inputType == -1 {
		return fmt.Errorf("an input source is required before an output source can be initialized")
	}

	if br.airplay == nil {
		br.airplay = newAirPlayHandler()
	}

	if br.airplay.clients == nil {
		br.airplay.clients = make([]*airplay2.Client, 0)
	}

	params := airplay2.ClientDiscoveryParameters{}

	switch searchType {
	case AirPlaySearchByIP:
		params.DeviceIP = searchKey
	case AirPlaySearchByName:
		params.DeviceNameRegex = searchKey
	default:
		return fmt.Errorf("invalid search type")
	}

	client, err := airplay2.NewClient(params)
	if err != nil {
		return fmt.Errorf("error initializing AirPlay client: %w", err)
	}

	// Close any connections that would be a duplicate of our current connection.
	for i := range br.airplay.clients {
		if br.airplay.clients[i].RemoteIP().Equal(client.RemoteIP()) {
			log.Logger.WithField("context", "AirPlay Client Init").Warnf("Closing previous session with matching remote address...")
			br.airplay.clients[i].Close()
		}
	}

	if err := client.ConfirmConnect(); err != nil {
		return fmt.Errorf("error confirming connection for AirPlay client: %w", err)
	}

	br.airplay.clients = append(br.airplay.clients, client)

	if err := br.wireAirPlayOutput(client); err != nil {
		return fmt.Errorf("error wiring AirPlay output to input: %w", err)
	}
	return nil
}

type AirPlayHandler struct {
	server  *airplay2.Server
	clients []*airplay2.Client
}

func newAirPlayHandler() *AirPlayHandler {
	return &AirPlayHandler{}
}

func (aph *AirPlayHandler) Stop() {
	if aph.clients != nil {
		for i := range aph.clients {
			aph.clients[i].Close()
		}
	}
	if aph.server != nil {
		aph.server.Stop()
	}
}

type AirPlaySearchType string

const (
	AirPlaySearchByName AirPlaySearchType = "name"
	AirPlaySearchByIP   AirPlaySearchType = "ip"
)
