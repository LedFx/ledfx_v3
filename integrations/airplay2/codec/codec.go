package codec

import (
	"strings"

	"github.com/carterpeel/bobcaygeon/rtsp"
	"github.com/maghul/alac"
)

// Handler is a function type for receiving raw bytes and decoding them using some codec
type Handler func(data []byte) ([]byte, error)

var codecMap = map[string]Handler{
	"AppleLossless": decodeAlac}

func decodeAlac(data []byte) ([]byte, error) {
	decoder, err := alac.New()
	if err != nil {
		return nil, err
	}
	return decoder.Decode(data), nil
}

// GetCodec determines the appropriate codec from the rtsp session
func GetCodec(session *rtsp.Session) Handler {
	var decoder Handler
	rtpmap := session.Description.Attributes["rtpmap"]
	if strings.Contains(rtpmap, "AppleLossless") {
		decoder = codecMap["AppleLossless"]
	} else {
		decoder = func(data []byte) ([]byte, error) { return data, nil }
	}
	return decoder
}
