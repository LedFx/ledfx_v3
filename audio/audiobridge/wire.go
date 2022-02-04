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
		br.airplay.server.AddClient(client)
	case inputTypeLocal:
		br.byteWriter.AppendWriter(client)
	default:
		err = fmt.Errorf("unrecognized input type '%d'", br.inputType)
	}
	return err
}

func (br *Bridge) wireLocalOutput(handler *playback.Handler) {
	br.byteWriter.AppendWriter(handler)
}

func (br *Bridge) AddOutputWriter(wr io.Writer) {
	br.byteWriter.AppendWriter(wr)
}
