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
		err = br.AddOutputWriter(client, client.WriterID())
	default:
		err = fmt.Errorf("unrecognized input type '%d'", br.inputType)
	}
	if err != nil {
		br.outputs = append(br.outputs, &OutputInfo{
			Type: outputTypeAirPlay,
			Info: &AirPlayOutputInfo{
				IP:          client.RemoteIP().String(),
				Hostname:    client.Hostname(),
				AdvertName:  client.Name(),
				Type:        client.Type(),
				Port:        client.RemotePort(),
				SampleRate:  client.SampleRate(),
				DeviceModel: client.DeviceModel(),
			},
		})
	}
	return err
}

func (br *Bridge) wireLocalOutput(handler playback.Handler) error {
	if err := br.AddOutputWriter(handler, handler.Identifier()); err != nil {
		return err
	}
	br.outputs = append(br.outputs, &OutputInfo{
		Type: outputTypeLocal,
		Info: &LocalOutputInfo{
			Device:     handler.Device(),
			Identifier: handler.Identifier(),
			SampleRate: handler.SampleRate(),
			Channels:   handler.NumChannels(),
		},
	})
	return nil
}

func (br *Bridge) AddOutputWriter(wr io.Writer, name string) error {
	return br.byteWriter.AddWriter(wr, name)
}
