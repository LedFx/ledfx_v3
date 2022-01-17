package airplay2

import (
	"github.com/carterpeel/bobcaygeon/player"
	"github.com/carterpeel/bobcaygeon/rtsp"
	"io"
	"ledfx/integrations/airplay2/codec"
	log "ledfx/logger"
	"sync"
	"sync/atomic"
)

type audioPlayer struct {
	volLock    sync.RWMutex
	volume     float64
	lastVolume float64

	outputMu sync.Mutex
	outputs  []io.Writer

	metaDataMu sync.RWMutex
	album      string
	artist     string
	title      string
	artwork    []byte

	muted *atomic.Value
}

func newPlayer() *audioPlayer {
	p := &audioPlayer{
		volLock:    sync.RWMutex{},
		outputMu:   sync.Mutex{},
		outputs:    make([]io.Writer, 0),
		metaDataMu: sync.RWMutex{},
		muted:      &atomic.Value{},
		volume:     1,
	}
	p.muted.Store(false)
	return p
}

func (p *audioPlayer) AddWriter(wr io.Writer) {
	p.outputMu.Lock()
	defer p.outputMu.Unlock()
	p.outputs = append(p.outputs, wr)
}

func (p *audioPlayer) Play(session *rtsp.Session) {
	go p.playStream(session)
}

func (p *audioPlayer) playStream(session *rtsp.Session) {
	// We need a writer waitgroup so the recieving player doesn't get confused
	// if one finishes before another
	wg := sync.WaitGroup{}

	decoder := codec.GetCodec(session)
	for d := range session.DataChan {
		p.volLock.RLock()
		vol := p.volume
		p.volLock.RUnlock()
		decoded, err := decoder(d)
		if err != nil {
			log.Logger.Warnf("Error decoding audio: %v\n", err)
			continue
		}
		adjusted := codec.AdjustAudio(decoded, vol)
		wg.Add(len(p.outputs))
		for _, output := range p.outputs {
			output := output
			go func() {
				_, _ = output.Write(adjusted)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func (p *audioPlayer) SetVolume(volume float64) {
	p.volLock.Lock()
	if volume <= 0 {
		p.SetMute(true)
	} else {
		p.muted.Store(false)
		p.volume = volume
	}
	p.volLock.Unlock()
}

func (p *audioPlayer) SetMute(isMuted bool) {
	p.volLock.Lock()
	switch isMuted {
	case true:
		p.muted.Store(true)
		p.lastVolume = p.volume
		p.volume = -1
	case false:
		p.muted.Store(false)
		p.volume = p.lastVolume
	}
	p.volLock.Unlock()
}

func (p *audioPlayer) GetIsMuted() bool {
	return p.muted.Load().(bool)
}

func (p *audioPlayer) SetTrack(album string, artist string, title string) {
	p.metaDataMu.Lock()
	p.album = album
	p.artist = artist
	p.title = title
	p.metaDataMu.Unlock()
}

func (p *audioPlayer) SetAlbumArt(artwork []byte) {
	p.metaDataMu.Lock()
	p.artwork = artwork
	p.metaDataMu.Unlock()
}

func (p *audioPlayer) GetTrack() player.Track {
	p.metaDataMu.RLock()
	defer p.metaDataMu.RUnlock()
	return player.Track{
		Artist:  p.artist,
		Album:   p.album,
		Title:   p.title,
		Artwork: p.artwork,
	}
}
