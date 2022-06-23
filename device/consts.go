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

type UDPProtocol string

const (
	WARLS UDPProtocol = "WARLS"
	DRGB  UDPProtocol = "DRGB"
	DRGBW UDPProtocol = "DRGBW"
	DNRGB UDPProtocol = "DNRGB"
	DDP   UDPProtocol = "DDP"
)

func (u UDPProtocol) Byte() byte {
	switch u {
	case WARLS:
		return byte(1)
	case DRGB:
		return byte(2)
	case DRGBW:
		return byte(3)
	case DNRGB:
		return byte(4)
	case DDP:
		return byte(5)
	default:
		return byte(0)
	}
}
