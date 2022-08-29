package sdp

// Origin section of a SDP payload
type Origin struct {
	Username       string
	SessionID      string
	SessionVersion string
	NetType        string
	AddrType       string
	UnicastAddress string
}

// ConnectData connection section of a SDP payload
type ConnectData struct {
	NetType           string
	AddrType          string
	ConnectionAddress string
}

// Timing timing section of a SDP payload
type Timing struct {
	StartTime int
	StopTime  int
}

// MediaDescription media description of a SDP payload
type MediaDescription struct {
	Media string
	Port  string // keeping string for now (parse later, fmt: <port>/<number of ports>)
	Proto string
	Fmt   string
}

// SessionDescription a struct representation of a SDP payload
type SessionDescription struct {
	Version          int
	Origin           Origin
	SessionName      string
	Information      string
	ConnectData      ConnectData
	Timing           Timing
	MediaDescription []MediaDescription
	Attributes       map[string]string
}

// NewSessionDescription instantiates a SessionDescription struct
func NewSessionDescription() *SessionDescription {
	var mediaDescription []MediaDescription
	sdp := SessionDescription{Version: 0, Attributes: make(map[string]string), MediaDescription: mediaDescription}
	return &sdp
}
