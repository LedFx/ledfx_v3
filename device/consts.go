package device

const (
	DDP_HEADER = 0x40 // ver 01
	DDP_PUSH   = 0x01 // push flag
	DDP_DTYPE  = 0x8A // RGB 8 bits per channel, though impl is undefined as per ver 01
	DDP_DEST   = 0x01 // default output device
)

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
	DNRGB UDPProtocol = 4
	DDP   UDPProtocol = 5
)

func (u UDPProtocol) String() string {
	switch u {
	case WARLS:
		return "WARLS"
	case DRGB:
		return "DRGB"
	case DRGBW:
		return "DRGBW"
	case DNRGB:
		return "DNRGB"
	case DDP:
		return "DDP"
	default:
		return ""
	}
}

var UDPProtocols = map[string]UDPProtocol{
	"WARLS": WARLS,
	"DRGB":  DRGB,
	"DRGBW": DRGBW,
	"DNRGB": DNRGB,
	"DDP":   DDP,
}
