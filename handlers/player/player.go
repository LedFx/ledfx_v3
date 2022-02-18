package player

import (
	"sync"

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
}

// LocalPlayer is a player that will just play the audio locally
type LocalPlayer struct {
	volLock sync.RWMutex
	volume  float64
}

// Track represents a track playing by the player
type Track struct {
	Artist  string
	Album   string
	Title   string
	Artwork []byte
}
