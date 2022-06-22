package audiobridge

import (
	"ledfx/audio/audiobridge/youtube"
)

func (br *Bridge) StartYoutubeInput() error {
	if br.inputType != -1 {
		br.closeInput()
	}

	br.inputType = inputTypeYoutube

	if br.youtube == nil {
		br.youtube = youtube.NewHandler(br.byteWriter)
	}
	return nil
}
