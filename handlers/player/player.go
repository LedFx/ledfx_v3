package player

import (
	"encoding/binary"
	"io"
	"log"
	"sync"

	"github.com/carterpeel/oto/v2"
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

// NewLocalPlayer instantiates a new LocalPlayer
func NewLocalPlayer() *LocalPlayer {
	return &LocalPlayer{volume: 1}
}

// Play will play the packets received on the specified session
func (lp *LocalPlayer) Play(session *rtsp.Session) {
	go lp.playStream(session)
}

// SetVolume accepts a float between 0 (mute) and 1 (full volume)
func (lp *LocalPlayer) SetVolume(volume float64) {
	lp.volLock.Lock()
	defer lp.volLock.Unlock()
	lp.volume = volume

}

// SetTrack sets the track for the player
func (lp *LocalPlayer) SetTrack(album string, artist string, title string) {
	// no op for now
}

// SetAlbumArt sets the album art for the player
func (lp *LocalPlayer) SetAlbumArt(artwork []byte) {
	// no op for now
}

// SetMute will mute or unmute the player
func (lp *LocalPlayer) SetMute(isMuted bool) {
	// no op for now
}

// GetIsMuted returns muted state
func (lp *LocalPlayer) GetIsMuted() bool {
	return false
}

// GetTrack returns the track
func (lp *LocalPlayer) GetTrack() Track {
	return Track{}
}

func (lp *LocalPlayer) playStream(session *rtsp.Session) {
	pctx, _, err := oto.NewContext(44100, 2, 2)
	if err != nil {
		log.Println("error initializing player", err)
		return
	}
	rd, wr := io.Pipe()

	p := pctx.NewPlayer(rd)
	p.Play(false)
	decoder := GetCodec(session)
	for d := range session.DataChan {
		lp.volLock.RLock()
		vol := lp.volume
		lp.volLock.RUnlock()
		decoded, err := decoder(d)
		if err != nil {
			log.Println("Problem decoding packet")
		}
		AdjustAudio(decoded, vol)
		wr.Write(decoded)
	}
	log.Println("Data stream ended closing player")
	p.Close()
}

// AdjustAudio takes a raw data frame of audio and a volume value between 0 and 1, 1 being full volume, 0 being mute
func AdjustAudio(raw []byte, vol float64) {
	if vol == 1 {
		return
	}
	for i := 0; i < len(raw); i += 2 {
		binary.LittleEndian.PutUint16(raw[i:i+2], uint16(max(-32767, min(32767, int16(vol*float64(int16(binary.LittleEndian.Uint16(raw[i:i+2]))))))))
	}
}
func min(a, b int16) int16 {
	if a < b {
		return a
	}
	return b
}
func max(a, b int16) int16 {
	if a > b {
		return a
	}
	return b
}
