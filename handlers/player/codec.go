package player

import (
	"strings"

	"github.com/maghul/alac"
	"ledfx/handlers/rtsp"
)

// CodecHandler handler function for receiving raw bytes and decoding them using some codec
type CodecHandler func(data []byte) ([]byte, error)

var codecMap = map[string]CodecHandler{
	"AppleLossless": decodeAlac}

func decodeAlac(data []byte) ([]byte, error) {
	decoder, err := alac.New()
	if err != nil {
		return nil, err
	}
	return decoder.Decode(data), nil
}

// GetCodec determines the appropriate codec from the rtsp session
func GetCodec(session *rtsp.Session) CodecHandler {
	var decoder CodecHandler
	rtpmap := session.Description.Attributes["rtpmap"]
	if strings.Contains(rtpmap, "AppleLossless") {
		decoder = codecMap["AppleLossless"]
	} else {
		decoder = func(data []byte) ([]byte, error) { return data, nil }
	}
	return decoder
}
