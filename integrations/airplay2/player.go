package airplay2

import (
	"github.com/dustin/go-broadcast"
	"io"
	"ledfx/audio"
	"ledfx/color"
	"ledfx/handlers/player"
	"ledfx/handlers/raop"
	"ledfx/handlers/rtsp"
	"ledfx/integrations/airplay2/codec"
	log "ledfx/logger"
	"unsafe"
)

type audioPlayer struct {
	/* Variables that are looped through often belong at the top of the struct */
	doEncodedSend, hasDecodedOutputs, sessionActive, muted bool

	numClients int
	apClients  [8]*Client

	// outputs supports 8 writers maximum.
	// We use a fixed-length array because
	// indexing an array is slightly faster
	// than indexing a slice.
	numOutputs int
	outputs    [8]io.Writer

	quit chan bool

	artwork []byte
	album   string
	artist  string
	title   string

	volume float64

	hermes broadcast.Broadcaster
}

func newPlayer(hermes broadcast.Broadcaster) *audioPlayer {
	p := &audioPlayer{
		outputs:   [8]io.Writer{},
		apClients: [8]*Client{},
		volume:    1,
		quit:      make(chan bool),
		hermes:    hermes,
	}
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
	p.hasDecodedOutputs = true
	p.outputs[p.numOutputs] = wr
	p.numOutputs++
}

func (p *audioPlayer) Play(session *rtsp.Session) {
	log.Logger.WithField("category", "AirPlay Player").Warnf("Starting new session")
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
				case !p.doEncodedSend && !p.hasDecodedOutputs:
					continue
				default:
					func() {
						defer func() {
							if err := recover(); err != nil {
								log.Logger.WithField("category", "AirPlay Player").Warnf("Recovered from panic during playStream: %v\n", err)
							}
						}()

						if p.doEncodedSend {
							p.broadcastEncoded(recvBuf)
						}

						recvBuf = dc.Decode(recvBuf)
						codec.NormalizeAudio(recvBuf, p.volume)

						p.hermes.Submit(audioBufFromBytes(recvBuf))

						if p.hasDecodedOutputs {
							for i := 0; i < p.numOutputs; i++ {
								_, _ = p.outputs[i].Write(recvBuf)
							}
						}
					}()
				}
			case <-p.quit:
				log.Logger.WithField("category", "AirPlay Player").Warnf("Session with peer '%s' closed", session.Description.ConnectData.ConnectionAddress)
				return
			}
		}
	}(decoder)
}

func audioBufFromBytes(recvBuf []byte) audio.Buffer {
	audioBuf := audio.Buffer{}
	var offset int
	for i := 0; i < len(recvBuf); i += 2 {
		audioBuf = append(audioBuf, readInt16Unsafe(recvBuf[i:i+2]))
		offset++
	}
	return audioBuf
}

func readInt16Unsafe(b []byte) int16 {
	return *(*int16)(unsafe.Pointer(&b[0]))
}

func (p *audioPlayer) AddClient(client *Client) {
	p.doEncodedSend = true
	p.apClients[p.numClients] = client
	p.numClients++
}

func (p *audioPlayer) SetVolume(volume float64) {
	p.volume = volume
	if p.doEncodedSend {
		p.broadcastParam(raop.ParamVolume(prepareVolume(volume)))
	}
	p.SetMute(volume == 0)
}

func (p *audioPlayer) SetMute(isMuted bool) {
	p.muted = isMuted
	if p.doEncodedSend {
		p.broadcastParam(raop.ParamMuted(isMuted))
	}
	if isMuted {
		log.Logger.WithField("category", "AirPlay Player").Infoln("Muting stream...")
	}
}

func (p *audioPlayer) GetIsMuted() bool {
	return p.muted
}

func (p *audioPlayer) SetTrack(album string, artist string, title string) {
	p.album = album
	p.artist = artist
	p.title = title
	if p.doEncodedSend {
		p.broadcastParam(raop.ParamTrackInfo{
			Album:  album,
			Artist: artist,
			Title:  title,
		})
	}
}

func (p *audioPlayer) SetAlbumArt(artwork []byte) {
	p.artwork = artwork
	if p.doEncodedSend {
		p.broadcastParam(raop.ParamAlbumArt(artwork))
	}
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
	if p.sessionActive {
		p.quit <- true
	}
}

func (p *audioPlayer) broadcastParam(par interface{}) {
	for i := range p.apClients {
		go p.apClients[i].SetParam(par)
	}
}

func (p *audioPlayer) broadcastEncoded(data []byte) {
	for i := 0; i < p.numClients; i++ {
		_, _ = p.apClients[i].DataConn.Write(data)
	}
}
