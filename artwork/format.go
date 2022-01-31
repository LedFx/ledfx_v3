package artwork

import (
	"bytes"
	// Side effects
	_ "golang.org/x/image/webp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"image"
	"mime"
)

// GuessImageFormat guesses the image format from gif/jpeg/png/webp
func GuessImageFormat(b []byte) (format string, err error) {
	_, format, err = image.DecodeConfig(bytes.NewReader(b))
	return
}

// GuessMimeTypes guesses image the mime types from gif/jpeg/png/webp
func GuessMimeTypes(b []byte) string {
	format, _ := GuessImageFormat(b)
	if format == "" {
		return ""
	}
	return mime.TypeByExtension("." + format)
}
