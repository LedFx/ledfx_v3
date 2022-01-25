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
	recvBuf    []byte
	encodedBuf []byte

	muted *atomic.Value
	//curSession *rtsp.Session

	closeChan chan struct{}

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
		closeChan:  make(chan struct{}),
		recvBuf:    make([]byte, 0),
		encodedBuf: make([]byte, 0),
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
	decoder := codec.GetCodec(session)
	go func(dc *codec.Handler) {
		for p.recvBuf = range session.DataChan {
			func() {
				defer func() {
					if err := recover(); err != nil {
						log.Logger.Errorf("Recovered from panic during playStream: %v\n", err)
					}
				}()
				if p.doEncodedSend {
					_, _ = p.apClient.DataConn.Write(p.recvBuf)
				}
				p.recvBuf = dc.Decode(p.recvBuf)
				codec.NormalizeAudio(p.recvBuf, p.volume)
				for i := 0; i < p.numOutputs; i++ {
					_, _ = p.outputs[i].Write(p.recvBuf)
				}
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
