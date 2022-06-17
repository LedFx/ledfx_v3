package youtube

import (
	"errors"
	"fmt"
	"io"
	"ledfx/audio"
	log "ledfx/logger"
	"ledfx/tickpool"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/atomic"
)

type Player struct {
	mu *sync.Mutex

	h *Handler

	isDone *atomic.Bool
	done   chan struct{}

	paused  *atomic.Bool
	pause   chan struct{}
	unpause chan struct{}

	playing *atomic.Bool
	play    chan []byte

	cycling *atomic.Bool

	next     chan struct{}
	nextDone chan struct{}
	prev     chan struct{}
	prevDone chan struct{}

	playByIndex     chan int32
	playByIndexDone chan struct{}

	in  *FileBuffer
	out *audio.AsyncMultiWriter

	elapsed *atomic.Duration
	ticker  *time.Ticker

	trackMu    *sync.Mutex
	trackNum   *atomic.Int32
	tracks     []TrackInfo
	trackPaths []string

	curWav *FileBuffer
}

func (p *Player) Download(URL string) error {
	parsed, err := url.Parse(URL)
	if err != nil {
		return fmt.Errorf("error parsing URL: %w", err)
	}
	URL = parsed.String()

	var toDownload []TrackInfo

	if strings.Contains(strings.ToLower(URL), "list=") {
		playlist, err := p.h.cl.GetPlaylist(URL)
		if err != nil {
			return fmt.Errorf("error getting playlist metadata from URL %q: %w", URL, err)
		}
		toDownload = make([]TrackInfo, len(playlist.Videos))

		for i := 0; i < len(playlist.Videos); i++ {
			entry := playlist.Videos[i]
			video, err := p.h.cl.VideoFromPlaylistEntry(playlist.Videos[i])
			if err != nil {
				log.Logger.WithField("context", "YouTube Download Handler").Errorf("Error getting entry metadata for %q: %v", entry.Title, err)
				playlist.Videos = append(playlist.Videos[:i], playlist.Videos[i+1:]...)
				toDownload = toDownload[:len(toDownload)-1]
				continue
			}
			toDownload[i] = TrackInfo{
				Artist:        entry.Author,
				Title:         entry.Title,
				Duration:      SongDuration(entry.Duration),
				SampleRate:    44100,
				FileSize:      -1,
				URL:           fmt.Sprintf("https://youtu.be/%s", entry.ID),
				AudioChannels: 2,
				video:         video,
			}
		}
	} else {
		video, err := p.h.cl.GetVideo(URL)
		if err != nil {
			return fmt.Errorf("error getting video metadata from URL %q: %w", URL, err)
		}
		toDownload = []TrackInfo{
			{
				Artist:        video.Author,
				Title:         video.Title,
				Duration:      SongDuration(video.Duration),
				SampleRate:    44100,
				FileSize:      -1,
				URL:           URL,
				AudioChannels: 2,
				video:         video,
			},
		}
	}

	p.trackMu.Lock()
	defer p.trackMu.Unlock()

	var clearBar bool

	for current, v := range toDownload {
		if current >= len(toDownload)-1 {
			clearBar = true
		}
		path, err := p.h.downloadWAV(v, current+1, len(toDownload), clearBar)
		if err != nil {
			log.Logger.WithField("context", "YouTube Download Handler").Errorf("Error downloading %q: %v", cleanString(v.Title), err)
			continue
		}
		p.tracks = append(p.tracks, v)
		p.trackPaths = append(p.trackPaths, path)
	}

	return nil
}

func (p *Player) Play() error {
	if len(p.trackPaths) <= 0 || len(p.tracks) <= 0 {
		return errors.New("no tracks found")
	}

	// This does nothing if the playback loop is already active.
	p.playLoop()

	// This does nothing if the track cycle is already active.
	p.cycleTracks()

	return nil
}

func (p *Player) Pause() {
	if p.playing.Load() && !p.paused.Load() {
		p.paused.Store(true)
		p.pause <- struct{}{}
	}
}

func (p *Player) Unpause() {
	if p.playing.Load() && p.paused.Load() {
		p.paused.Store(false)
		p.unpause <- struct{}{}
	}
}

func (p *Player) Next() {
	if p.playing.Load() {
		if p.paused.Load() {
			p.Unpause()
		}
		p.next <- struct{}{}
		<-p.nextDone
	}
}
func (p *Player) Previous() {
	if p.playing.Load() {
		if p.paused.Load() {
			p.Unpause()
		}
		p.prev <- struct{}{}
		<-p.prevDone
	}
}

func (p *Player) IsPlaying() bool {
	return p.playing.Load()
}

func (p *Player) IsPaused() bool {
	return p.paused.Load()
}

func (p *Player) PlayTrack(index int) error {
	switch {
	case index >= len(p.tracks), index < 0:
		return fmt.Errorf("index value must be between 0 and %d", len(p.tracks))
	case !p.playing.Load():
		if err := p.Play(); err != nil {
			return err
		}
	case p.paused.Load():
		p.Unpause()
	}

	p.playByIndex <- int32(index)
	<-p.playByIndexDone

	return nil
}

func (p *Player) PlayTrackByName(name string) error {
	for index := range p.tracks {
		if strings.Contains(strings.ToLower(p.tracks[index].Title), strings.ToLower(name)) {
			return p.PlayTrack(index)
		}
	}
	return fmt.Errorf("cannot find track with name %q", name)
}

func (p *Player) NowPlaying() TrackInfo {
	if p.cycling.Load() {
		p.trackMu.Lock()
		defer p.trackMu.Unlock()
		return p.tracks[p.trackNum.Load()]
	} else {
		return TrackInfo{
			Artist:        "N/A",
			Title:         "N/A",
			Duration:      0,
			SampleRate:    -1,
			FileSize:      -1,
			URL:           "N/A",
			AudioChannels: -1,
			Invalid:       true,
		}
	}
}

func (p *Player) QueuedTracks() []TrackInfo {
	p.trackMu.Lock()
	defer p.trackMu.Unlock()
	return p.tracks
}

func (p *Player) Close() error {
	if p.playing.Load() {
		p.done <- struct{}{}
	} else {
		return errors.New("cannot close inactive player")
	}

	return nil
}

func (p *Player) TimeElapsed() time.Duration {
	return p.elapsed.Load()
}

func (p *Player) TrackIndex() int {
	return int(p.trackNum.Load())
}

func (p *Player) elapsedLoop(done chan struct{}) {
	p.ticker = tickpool.Get(time.Second)
	defer tickpool.Put(p.ticker)

	for {
		select {
		case <-done:
			return
		case <-p.ticker.C:
			if !p.paused.Load() {
				p.elapsed.Add(1 * time.Second)
			}
		}
	}
}

func (p *Player) playLoop() {
	if p.playing.Load() {
		return
	}

	p.playing.Store(true)

	go func() {
		defer func() {
			p.playing.Store(false)
		}()

		for {
			select {
			case index := <-p.playByIndex:
				p.trackMu.Lock()
				p.trackNum.Store(index)
				_ = p.curWav.Close()
				p.elapsed.Store(0)
				p.trackMu.Unlock()

				p.playByIndexDone <- struct{}{}
			case <-p.next:
				p.trackMu.Lock()

				if int(p.trackNum.Inc()) >= len(p.trackPaths) {
					p.trackNum.Store(0)
				}
				_ = p.curWav.Close()
				p.elapsed.Store(0)

				p.trackMu.Unlock()

				p.nextDone <- struct{}{}
			case <-p.prev:
				p.trackMu.Lock()

				if p.trackNum.Dec() < 0 {
					p.trackNum.Store(int32(len(p.trackPaths) - 1))
				}
				_ = p.curWav.Close()
				p.elapsed.Store(0)

				p.trackMu.Unlock()

				p.prevDone <- struct{}{}
			case <-p.pause:
				<-p.unpause
			case <-p.done:
				log.Logger.WithField("context", "YouTube Playback Loop").Warnf("Got 'DONE' signal...")
				return
			case buf := <-p.play:
				_, _ = p.out.Write(buf)
			}
		}
	}()
}

func (p *Player) cycleTracks() {
	if p.cycling.Load() {
		return
	}

	p.cycling.Store(true)
	p.trackNum = atomic.NewInt32(0)

	go func() {
		elapsedDone := make(chan struct{})
		go p.elapsedLoop(elapsedDone)
		defer func() {
			close(p.play)
			close(p.playByIndex)
			close(p.playByIndexDone)
			close(p.pause)
			close(p.unpause)
			close(p.done)
			close(elapsedDone)
		}()

		buf := make([]byte, 1408)
		for p.playing.Load() {
			p.trackMu.Lock()
			path := p.trackPaths[p.trackNum.Load()]
			p.trackMu.Unlock()

			func() {
				wav, err := os.Open(path)
				if err != nil {
					log.Logger.WithField("context", "YouTube Track Cycle").Errorf("Error opening %q: %v", path, err)
					return
				}

				if p.curWav, err = NewFileBuffer(wav); err != nil {
					log.Logger.WithField("context", "YouTube Track Cycle").Errorf("Error creating new file buffer: %v", err)
					return
				}
				defer p.curWav.Close()

				for {
					n, err := p.curWav.Read(buf)
					if err != nil {
						switch {
						case errors.Is(err, os.ErrClosed):
							return
						case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
							p.Next()
							if n > 0 {
								goto Send
							} else {
								return
							}
						}
					}
				Send:
					cpy := make([]byte, n)
					copy(cpy, buf[:n])
					p.play <- cpy
				}

			}()

		}

	}()

}
