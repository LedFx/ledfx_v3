package codec

import (
	"strings"

	alac "github.com/carterpeel/go.alac"
	"ledfx/handlers/rtsp"
)

// Handler is a function type for receiving raw bytes and decoding them using some codec
type Handler struct {
	decoderFn func(data []byte) []byte
	a         *alac.Alac
}

func (h *Handler) Free() {
	h.a = nil
}

func (h *Handler) Decode(in []byte) []byte {
	return h.decoderFn(in)
}

func GetCodec(session *rtsp.Session) (decoder *Handler) {
	rtpmap := session.Description.Attributes["rtpmap"]
	if strings.Contains(rtpmap, "AppleLossless") {
		a, _ := alac.New()
		decoder = &Handler{
			decoderFn: func(data []byte) []byte { return a.Decode(data) },
			a:         a,
		}
	} else {
		decoder = &Handler{
			decoderFn: func(data []byte) []byte { return data },
		}
	}
	return decoder
}
