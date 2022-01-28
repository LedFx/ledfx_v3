package audiobridge

import (
	"ledfx/color"
)

func (br *Bridge) GetGradientFromArtwork(resolution int) (*color.Gradient, error) {
	if br.airplayServer != nil {
		return br.airplayServer.GetAlbumGradient(resolution)
	}
	return nil, ErrAirplayRequiredForArtworkGradient
}
