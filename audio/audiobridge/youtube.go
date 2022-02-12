package audiobridge

import (
	"fmt"
	"ledfx/audio/audiobridge/youtube"
)

type YoutubeHandler struct {
	handler *youtube.Handler
}

func (br *Bridge) StartYoutubeInput(verbose bool) error {
	if br.inputType != -1 {
		return fmt.Errorf("an input source has already been defined for this bridge")
	}
	br.inputType = inputTypeYoutube

	if br.youtube == nil {
		br.youtube = &YoutubeHandler{
			handler: youtube.NewHandler(br.intWriter, br.byteWriter, verbose),
		}
	}
	return nil
}
