package airplay2

import (
	"encoding/json"
	"fmt"
	"ledfx/audio"
	"ledfx/handlers/player"
	"ledfx/handlers/raop"
	"ledfx/handlers/rtsp"
	"ledfx/integrations/airplay2/codec"
	log "ledfx/logger"
	"sync"
	"unsafe"
)

type audioPlayer struct {
	/* Variables that are looped through often belong at the top of the struct */
	wg sync.WaitGroup

	byteWriter *audio.AsyncMultiWriter

	hasClients, sessionActive, muted bool

	numClients int
	apClients  []*Client

	quit chan bool

	artwork []byte

	title  string
	artist string
	album  string

	volume float64
}

func (p *audioPlayer) MarshalJSON() (b []byte, err error) {
	return json.Marshal(&struct {
		Title  string `json:"title"`
		Artist string `json:"artist"`
		Album  string `json:"album"`

		Volume float64 `json:"volume"`

		HasClients    bool `json:"has_clients"`
		NumClients    int  `json:"num_clients"`
		SessionActive bool `json:"session_active"`
		Muted         bool `json:"muted"`
	}{
		Title:         p.title,
		Artist:        p.artist,
		Album:         p.album,
		Volume:        p.volume,
		HasClients:    p.hasClients,
		NumClients:    p.numClients,
		SessionActive: p.sessionActive,
		Muted:         p.muted,
	})
}
func newPlayer(byteWriter *audio.AsyncMultiWriter) *audioPlayer {
	p := &audioPlayer{
		apClients:  make([]*Client, 0),
		volume:     1,
		quit:       make(chan bool),
		wg:         sync.WaitGroup{},
		byteWriter: byteWriter,
	}

	return p
}

func (p *audioPlayer) Play(session *rtsp.Session) {
	log.Logger.WithField("context", "AirPlay Player").Warnf("Starting new session")
	p.sessionActive = true
	decoder := codec.GetCodec(session)
	go func(dc *codec.Handler) {
		defer func() {
			p.sessionActive = false
		}()
		for {
			select {
			case recvBuf, ok := <-session.DataChan:
				switch {
				case !ok:
					return
				case p.muted:
					continue
				default:
					func() {
						defer func() {
							if err := recover(); err != nil {
								log.Logger.WithField("context", "AirPlay Player").Errorf("Recovered from panic during playStream: %v\n", err)
							}
						}()

						recvBuf = dc.Decode(recvBuf)
						codec.NormalizeAudio(recvBuf, p.volume)

						if _, err := p.byteWriter.Write(recvBuf); err != nil {
							log.Logger.WithField("context", "AirPlay Player").Errorf("Error writing to byteWriter: %v", err)
						}
					}()
				}
			case <-p.quit:
				log.Logger.WithField("context", "AirPlay Player").Warnf("Session with peer '%s' closed", session.Description.ConnectData.ConnectionAddress)
				return
			}
		}
	}(decoder)
}

func bytesToAudioBufferUnsafe(p []byte) (out audio.Buffer) {
	out = make([]int16, len(p))
	var offset int
	for i := 0; i < len(p); i += 2 {
		out[offset] = twoBytesToInt16Unsafe(p[i : i+2])
		offset++
	}
	return
}
func twoBytesToInt16Unsafe(p []byte) (out int16) {
	return *(*int16)(unsafe.Pointer(&p[0]))
}

func (p *audioPlayer) AddClient(client *Client) (err error) {
	p.hasClients = true
	p.apClients = append(p.apClients, client)
	if err := p.byteWriter.AddWriter(p.apClients[p.numClients], client.WriterID()); err != nil {
		return fmt.Errorf("error adding writer: %w", err)
	}
	p.numClients++
	return nil
}

func (p *audioPlayer) SetVolume(volume float64) {
	p.volume = volume
	if p.hasClients {
		p.broadcastParam(raop.ParamVolume(prepareVolume(volume)))
	}
}

func (p *audioPlayer) SetMute(isMuted bool) {
	p.muted = isMuted
	if p.hasClients {
		p.broadcastParam(raop.ParamMuted(isMuted))
	}
	if isMuted {
		log.Logger.WithField("context", "AirPlay Player").Infoln("Muting stream...")
	}
}

func (p *audioPlayer) GetIsMuted() bool {
	return p.muted
}

func (p *audioPlayer) SetTrack(album string, artist string, title string) {
	p.album = album
	p.artist = artist
	p.title = title
	if p.hasClients {
		p.broadcastParam(raop.ParamTrackInfo{
			Album:  album,
			Artist: artist,
			Title:  title,
		})
	}
}

func (p *audioPlayer) SetAlbumArt(artwork []byte) {
	p.artwork = artwork
	if p.hasClients {
		p.broadcastParam(raop.ParamAlbumArt(artwork))
	}
}

func (p *audioPlayer) GetTrack() player.Track {
	return player.Track{
		Artist:  p.artist,
		Album:   p.album,
		Title:   p.title,
		Artwork: p.artwork,
	}
}

func (p *audioPlayer) GetAlbumArt() []byte {
	return p.artwork
}

// airplay server will apply a normalization,
// we have the raw volume on a scale of 0 to 1,
// so we build the proper format. (-144 through 0)
func prepareVolume(vol float64) float64 {
	switch {
	case vol == 0:
		return -144
	case vol == 1:
		return 0
	default:
		return (vol * 30) - 30
	}
}

func (p *audioPlayer) Close() {
	if p.sessionActive {
		p.quit <- true
	}
}

func (p *audioPlayer) broadcastParam(par interface{}) {
	for i := 0; i < p.numClients; i++ {
		p.apClients[i].SetParam(par)
	}
}
