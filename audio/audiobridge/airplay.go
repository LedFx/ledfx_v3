package audiobridge

import (
	"fmt"
	"github.com/dustin/go-broadcast"
	"ledfx/audio"
	"ledfx/integrations/airplay2"
)

func (br *Bridge) StartAirPlayInput(name string, port int, verbose bool) error {
	if br.inputType != -1 {
		return fmt.Errorf("an input source has already been defined for this bridge")
	}
	br.inputType = inputTypeAirPlayServer

	if br.airplay == nil {
		br.airplay = newAirPlayHandler(br.hermes)
	}

	if br.airplay.server == nil {
		br.airplay.server = airplay2.NewServer(airplay2.Config{
			AdvertisementName: name,
			Port:              port,
			VerboseLogging:    verbose,
		}, br.local.hermes)
	}

	go func() {
		for audioFrame := range br.local.hermesChan {
			br.bufferCallback(audioFrame.(audio.Buffer))
		}
	}()

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
		br.airplay = newAirPlayHandler(br.hermes)
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
	server     *airplay2.Server
	clients    []*airplay2.Client
	hermes     broadcast.Broadcaster
	hermesChan chan interface{}
}

func newAirPlayHandler(hermes broadcast.Broadcaster) *AirPlayHandler {
	a := &AirPlayHandler{
		hermes:     hermes,
		hermesChan: make(chan interface{}),
	}
	a.hermes.Register(a.hermesChan)
	return a
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
