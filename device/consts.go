package device

type State int

const (
	Disconnected State = iota // default
	Connected
	Disconnecting
	Connecting
)

func (s State) String() string {
	switch s {
	case Connected:
		return "connected"
	case Disconnected:
		return "disconnected"
	case Connecting:
		return "connecting"
	case Disconnecting:
		return "disconnecting"
	}
	return "unknown" // this wont happen
}

type UDPProtocol byte

const (
	WARLS UDPProtocol = 1
	DRGB  UDPProtocol = 2
	DRGBW UDPProtocol = 3
)

func (u UDPProtocol) String() string {
	switch u {
	case WARLS:
		return "WARLS"
	case DRGB:
		return "DRGB"
	case DRGBW:
		return "DRGBW"
	default:
		return ""
	}
}

var UDPProtocols = map[string]UDPProtocol{
	"WARLS": WARLS,
	"DRGB":  DRGB,
	"DRGBW": DRGBW,
}
