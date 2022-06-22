package audiobridge

import (
	"ledfx/audio/audiobridge/youtube"
)

type YoutubeHandler struct {
	handler *youtube.Handler
}

func (br *Bridge) StartYoutubeInput() error {
	if br.inputType != -1 {
		br.closeInput()
	}

	br.inputType = inputTypeYoutube

	if br.youtube == nil {
		br.youtube = &YoutubeHandler{
			handler: youtube.NewHandler(br.byteWriter),
		}
	}
	return nil
}
