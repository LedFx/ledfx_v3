package airplay2

import (
	"io"
	"ledfx/color"
	"ledfx/handlers/player"
	"ledfx/handlers/raop"
	"ledfx/handlers/rtsp"
	"ledfx/integrations/airplay2/codec"
	log "ledfx/logger"
	"sync/atomic"
)

type audioPlayer struct {
	recvBuf    []byte
	encodedBuf []byte

	muted *atomic.Value

	quit       chan struct{}
	curSession *rtsp.Session

	// outputs supports 8 writers maximum.
	// We use a fixed-length array because
	// indexing an array is slightly faster
	// than indexing a slice.
	numOutputs int
	outputs    [8]io.Writer

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
		outputs:    [8]io.Writer{},
		muted:      &atomic.Value{},
		volume:     1,
		recvBuf:    make([]byte, 0),
		encodedBuf: make([]byte, 0),
		quit:       make(chan struct{}),
	}
	p.muted.Store(false)

	return p
}

// AddWriter adds a writer to p.output, which is wrapped with an io.MultiWriter().
//
//
// DO NOT provide an io.Writer() wrapped with io.MultiWriter()!
//
//------------------------------------------------------------
//
// If an io.MultiWriter() is provided, the already mediocre time complexity
// of this operation will go from O(n) to O(n^2). We don't want that.
func (p *audioPlayer) AddWriter(wr io.Writer) {
	p.outputs[p.numOutputs] = wr
	p.numOutputs++
}

func (p *audioPlayer) Play(session *rtsp.Session) {
	log.Logger.WithField("category", "AirPlay Player").Warnf("Starting new session")
	p.curSession = session
	decoder := codec.GetCodec(session)
	go func(dc *codec.Handler) {
		var ok bool
		for {
			select {
			case p.recvBuf, ok = <-session.DataChan:
				if !ok {
					return
				}
				func() {
					defer func() {
						if err := recover(); err != nil {
							log.Logger.WithField("category", "AirPlay Player").Warnf("Recovered from panic during playStream: %v\n", err)
						}
					}()
					if p.doEncodedSend {
						_, _ = p.apClient.DataConn.Write(p.recvBuf)
					}
					if p.numOutputs > 0 {
						p.recvBuf = dc.Decode(p.recvBuf)
						codec.NormalizeAudio(p.recvBuf, p.volume)
						for i := 0; i < p.numOutputs; i++ {
							_, _ = p.outputs[i].Write(p.recvBuf)
						}
					}
				}()
			case <-p.quit:
				log.Logger.WithField("category", "AirPlay Player").Warnf("Session with peer '%s' closed", session.Description.ConnectData.ConnectionAddress)
				return
			}
		}
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

func (p *audioPlayer) GetGradientFromArtwork(resolution int) (*color.Gradient, error) {
	return color.GradientFromPNG(p.artwork, resolution, 75)
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

func (p *audioPlayer) Close() {
	if p.curSession != nil {
		p.quit <- struct{}{}
	}
}
