package audiobridge

import (
	"fmt"
	"ledfx/integrations/airplay2"
)

func (br *Bridge) StartAirPlayInput(name string, port int, verbose bool) error {
	if br.inputType != -1 {
		return fmt.Errorf("an input source has already been defined for this bridge")
	}
	br.inputType = inputTypeAirPlayServer

	if br.airplay == nil {
		br.airplay = new(AirPlayHandler)
	}

	br.airplay.server = airplay2.NewServer(airplay2.Config{
		AdvertisementName: name,
		Port:              port,
		VerboseLogging:    verbose,
	})

	br.airplay.server.AddOutput(br.ledFxWriter)

	if err := br.airplay.server.Start(); err != nil {
		return fmt.Errorf("error starting AirPlay server: %w", err)
	}
	return nil
}

func (br *Bridge) AddAirPlayOutput(searchKey string, searchType AirPlaySearchType, verbose bool) error {
	if br.inputType == -1 {
		return fmt.Errorf("an input source is required before an output source can be initialized")
	}

	if br.airplay == nil {
		br.airplay = new(AirPlayHandler)
	}

	if br.airplay.clients == nil {
		br.airplay.clients = make([]*airplay2.Client, 0)
	}

	params := airplay2.ClientDiscoveryParameters{
		Verbose: verbose,
	}

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

type AirPlaySearchType int8

const (
	AirPlaySearchByName AirPlaySearchType = iota
	AirPlaySearchByIP
)
