package audiobridge

import "fmt"

func (br *Bridge) Info() *Info {
	return br.info
}

type Info struct {
	br *Bridge
}

func (i *Info) InputType() string {
	switch i.br.inputType {
	case inputTypeYoutube:
		return "youtube"
	case inputTypeLocal:
		return "local_capture"
	case inputTypeAirPlayServer:
		return "airplay_server"
	case -1:
		return "unspecified"
	default:
		return fmt.Sprintf("unknown (%d)", i.br.inputType)
	}
}

func (i *Info) AllOutputs() []*OutputInfo {
	return i.br.outputs
}
