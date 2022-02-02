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
		br.local.capture.AddByteWriters(client.DataConn)
	default:
		err = fmt.Errorf("unrecognized input type '%d'", br.inputType)
	}
	return err
}

func (br *Bridge) wireLocalOutput(playbackHandler *playback.Handler) (err error) {
	switch br.inputType {
	case -1:
		err = fmt.Errorf("input source has not been defined")
	case inputTypeAirPlayServer:
		br.airplay.server.AddOutput(playbackHandler)
	case inputTypeLocal:
		br.local.capture.AddByteWriters(playbackHandler)
	default:
		err = fmt.Errorf("unrecognized input type '%d'", br.inputType)
	}
	return err
}

func (br *Bridge) AddOutputWriter(wr io.Writer) (err error) {
	switch br.inputType {
	case -1:
		err = fmt.Errorf("input source has not been defined")
	case inputTypeAirPlayServer:
		br.airplay.server.AddOutput(wr)
	case inputTypeLocal:
		br.local.capture.AddByteWriters(wr)
	default:
		err = fmt.Errorf("unrecognized input type '%d'", br.inputType)
	}
	return err
}
