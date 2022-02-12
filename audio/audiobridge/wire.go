package audiobridge

import (
	"fmt"
	"io"
	"ledfx/audio/audiobridge/playback"
	"ledfx/integrations/airplay2"
)

func (br *Bridge) wireAirPlayOutput(client *airplay2.Client) (err error) {
	switch br.inputType {
	case -1:
		err = fmt.Errorf("input source has not been defined")
	case inputTypeAirPlayServer:
		err = br.airplay.server.AddClient(client)
	case inputTypeLocal:
		fallthrough
	case inputTypeYoutube:
		err = br.AddOutputWriter(client, client.Identifier())
	default:
		err = fmt.Errorf("unrecognized input type '%d'", br.inputType)
	}
	return err
}

func (br *Bridge) wireLocalOutput(handler *playback.Handler) error {
	return br.AddOutputWriter(handler, handler.Identifier())
}

func (br *Bridge) AddOutputWriter(wr io.Writer, name string) error {
	return br.byteWriter.AddWriter(wr, name)
}
