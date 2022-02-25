package player

import (
	"ledfx/handlers/rtsp"
)

// Player defines a player for outputting the data packets from the session
type Player interface {
	Play(session *rtsp.Session)
	SetVolume(volume float64)
	SetMute(isMuted bool)
	GetIsMuted() bool
	SetTrack(album string, artist string, title string)
	SetAlbumArt(artwork []byte)
	GetTrack() Track
	GetAlbumArt() []byte
}

// Track represents a track playing by the player
type Track struct {
	Artist  string
	Album   string
	Title   string
	Artwork []byte
}
