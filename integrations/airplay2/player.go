package airplay2

import (
	"github.com/carterpeel/bobcaygeon/player"
	"github.com/carterpeel/bobcaygeon/raop"
	"github.com/carterpeel/bobcaygeon/rtsp"
	"io"
	"ledfx/integrations/airplay2/codec"
	log "ledfx/logger"
	"sync/atomic"
)

type audioPlayer struct {
	muted *atomic.Value
	//curSession *rtsp.Session

	closeChan chan struct{}

	output io.Writer

	apClient      *Client
	doEncodedSend bool

	volume  float64
	artwork []byte

	album  string
	artist string
	title  string
}

func newPlayer() *audioPlayer {
	p := &audioPlayer{
		output:    io.MultiWriter(),
		muted:     &atomic.Value{},
		volume:    1,
		closeChan: make(chan struct{}),
	}
	p.muted.Store(false)

	return p
}

func (p *audioPlayer) AddWriter(wr io.Writer) {
	p.output = io.MultiWriter(p.output, wr)
}

func (p *audioPlayer) Play(session *rtsp.Session) {
	decoder := codec.GetCodec(session)
	go func(dc *codec.Handler) {
		for d := range session.DataChan {
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Logger.Errorf("Recovered from panic during playStream: %v\n", err)
					}
				}()
				if p.doEncodedSend {
					_, _ = p.apClient.Write(d)
				}
				buf := dc.Decode(d)
				codec.NormalizeAudio(buf, p.volume)
				_, _ = p.output.Write(buf)
			}()
		}
		log.Logger.Warnf("Session '%s' closed", session.Description.SessionName)
	}(decoder)
}

func (p *audioPlayer) SetClient(client *Client) {
	p.apClient = client
	p.doEncodedSend = true
}

func (p *audioPlayer) SetVolume(volume float64) {
	if p.doEncodedSend {
		p.apClient.SetParam(raop.ParamVolume(prepareVolume(volume)))
	}
	p.volume = volume
}

func (p *audioPlayer) SetMute(isMuted bool) {
	if p.doEncodedSend {
		p.apClient.SetParam(raop.ParamMuted(isMuted))
	}
	p.muted.Store(isMuted)
}

func (p *audioPlayer) GetIsMuted() bool {
	return p.muted.Load().(bool)
}

func (p *audioPlayer) SetTrack(album string, artist string, title string) {
	if p.doEncodedSend {
		p.apClient.SetParam(raop.ParamTrackInfo{
			Album:  album,
			Artist: artist,
			Title:  title,
		})
	}
	p.album = album
	p.artist = artist
	p.title = title
}

func (p *audioPlayer) SetAlbumArt(artwork []byte) {
	if p.doEncodedSend {
		p.apClient.SetParam(raop.ParamAlbumArt(artwork))
	}
	p.artwork = artwork
}

func (p *audioPlayer) GetTrack() player.Track {
	return player.Track{
		Artist:  p.artist,
		Album:   p.album,
		Title:   p.title,
		Artwork: p.artwork,
	}
}

// airplay server will apply a normalization,
// we have the raw volume on a scale of 0 to 1,
// so we build the proper format. (-144 through 0)
func prepareVolume(vol float64) float64 {
	switch vol {
	case 0:
		return -144
	case 1:
		return 0
	default:
		return (vol * 30) - 30
	}
}
