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

type Protocol string

const (
	WARLS Protocol = "WARLS" // https://github.com/Aircoookie/WLED/wiki/UDP-Realtime-Control
	DRGB  Protocol = "DRGB"
	DRGBW Protocol = "DRGBW"
	DNRGB Protocol = "DNRGB"
	DDP   Protocol = "DDP"      // http://www.3waylabs.com/ddp/
	ADA   Protocol = "Adalight" // https://gist.github.com/tvdzwan/9008833#file-adalightws2812-ino
	TPM2  Protocol = "TPM2"     // https://gist.github.com/jblang/89e24e2655be6c463c56
)
