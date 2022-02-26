package youtube

import (
	"fmt"
	"io"
	"ledfx/audio"
	log "ledfx/logger"
	"ledfx/util"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"
	pretty "github.com/fatih/color"
	yt "github.com/kkdai/youtube/v2"
	progressbar "github.com/schollz/progressbar/v3"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.uber.org/atomic"
)

type Handler struct {
	cl         *yt.Client
	byteWriter *audio.AsyncMultiWriter
	verbose    bool
	p          *Player
	pp         *PlaylistPlayer
	stopped    bool

	nowPlaying TrackInfo
	history    []TrackInfo
}

func (h *Handler) Quit() {
	h.stopped = true
	if h.p != nil {
		h.p.Close()
	}
	if h.pp != nil {
		h.pp.Stop()
	}
}
func (h *Handler) Stopped() bool {
	return h.stopped
}

func NewHandler(byteWriter *audio.AsyncMultiWriter, verbose bool) *Handler {
	h := &Handler{
		cl: &yt.Client{
			Debug:      false,
			HTTPClient: http.DefaultClient,
		},
		byteWriter: byteWriter,
		verbose:    verbose,
		history:    make([]TrackInfo, 0),
		p: &Player{
			mu:      &sync.Mutex{},
			done:    atomic.NewBool(false),
			paused:  atomic.NewBool(false),
			unpause: make(chan bool),
			playing: atomic.NewBool(false),
			in:      nil,
			out:     byteWriter,
			elapsed: atomic.NewDuration(0),
		},
	}
	h.pp = &PlaylistPlayer{
		h:        h,
		trackNum: -1,
		tracks:   make([]TrackInfo, 0),
	}
	return h
}
func (h *Handler) Play(url string) (p *Player, err error) {
	trackInfo, tmp, err := h.downloadToMP3(url)
	if err != nil {
		return nil, fmt.Errorf("error writing video data to tmpfile: %w", err)
	}

	h.history = append(h.history, trackInfo)
	h.nowPlaying = trackInfo

	if h.verbose {
		log.Logger.WithField("category", "YT Player").Infof(
			"[ARTIST=\"%s\", SAMPLERATE=\"%d\", SIZE=\"%s\", CHANNELS=\"%d\", DURATION=\"%s\"]",
			trackInfo.Artist,
			trackInfo.SampleRate,
			humanize.Bytes(uint64(trackInfo.FileSize)),
			trackInfo.AudioChannels,
			time.Duration(trackInfo.Duration).String(),
		)
	}

	logTrack(trackInfo.Title, trackInfo.Artist)

	fileBuffer, err := NewFileBuffer(tmp)
	if err != nil {
		return nil, fmt.Errorf("error creating new file buffer: %v", err)
	}

	h.p.Reset(fileBuffer)

	return h.p, nil
}

func (h *Handler) NowPlaying() TrackInfo {
	return h.nowPlaying
}
func (h *Handler) QueuedTracks() []TrackInfo {
	return h.pp.tracks
}
func (h *Handler) TimeElapsed() time.Duration {
	return h.p.elapsed.Load()
}

func (h *Handler) IsPaused() bool {
	return h.p.paused.Load()
}
func (h *Handler) TrackIndex() int {
	if len(h.QueuedTracks()) == 0 {
		return 0
	} else {
		return h.pp.trackNum
	}
}
func (h *Handler) IsPlaying() bool {
	return h.p.IsPlaying()
}

type CompletionPercent float32

func (c CompletionPercent) MarshalJSON() ([]byte, error) {
	if c > 100 || c < 0 {
		return nil, fmt.Errorf("CompletionPercent can only be in range 0-100")
	}
	return []byte(fmt.Sprintf("%0.2f", c)), nil
}

func (h *Handler) PercentComplete() (CompletionPercent, error) {
	if !h.IsPlaying() {
		return 0, nil
	}
	return CompletionPercent((float32(h.p.in.CurrentOffset()) / float32(h.NowPlaying().FileSize)) * float32(100)), nil
}

func (h *Handler) PlayPlaylist(playlistURL string) (pp *PlaylistPlayer, err error) {
	pl, err := h.cl.GetPlaylist(playlistURL)
	if err != nil {
		return nil, fmt.Errorf("error getting playlist: %w", err)
	}

	h.pp.Stop()

	h.pp.tracks = make([]TrackInfo, len(pl.Videos))

	sentReady := atomic.NewBool(false)
	ready := make(chan struct{})

	go func(r chan struct{}, sentR *atomic.Bool) {
		for i := range pl.Videos {
			h.pp.tracks[i] = TrackInfo{
				Artist:   pl.Videos[i].Author,
				Title:    pl.Videos[i].Title,
				Duration: SongDuration(pl.Videos[i].Duration),
				URL:      fmt.Sprintf("https://youtu.be/%s", pl.Videos[i].ID),
			}

			_, tmp, err := h.downloadToMP3(h.pp.tracks[i].URL)
			if err != nil {
				if tmp != nil {
					_ = tmp.Close()
				}
				log.Logger.WithField("category", "YouTube Downloader").Errorf("Error downloading playlist entry %d as MP3: %v", i, err)
				h.pp.tracks[i].invalid = true
				continue
			}
			_ = tmp.Close()
			if !sentR.Load() {
				sentR.Store(true)
				r <- struct{}{}
			}
		}
	}(ready, sentReady)

	go func(r chan struct{}) {
		defer close(r)
		<-r
		for {
			if err := h.pp.Next(true); err != nil {
				log.Logger.WithField("category", "Playlist Player").Errorf("Error waiting for next song to complete: %v", err)
				return
			}
		}
	}(ready)

	return h.pp, nil
}

func (h *Handler) downloadToMP3(url string) (videoInfo TrackInfo, tmp *os.File, err error) {
	video, err := h.cl.GetVideo(url)
	if err != nil {
		return videoInfo, nil, fmt.Errorf("error getting video: %w", err)
	}

	videoInfo = TrackInfo{
		Artist:   video.Author,
		Title:    video.Title,
		Duration: SongDuration(video.Duration),
		URL:      url,
	}

	audioFile := filepath.Join(os.TempDir(), cleanTitle(video.Title)+".wav")
	tmpVideoName := fmt.Sprintf("%s.ytdl", util.RandString(8))
	tmpVideoNameAndPath := filepath.Join(os.TempDir(), tmpVideoName)

	if util.FileExists(audioFile) {
		if tmp, err = os.Open(audioFile); err != nil {
			goto Download
		}
		st, err := tmp.Stat()
		if err != nil {
			return videoInfo, nil, fmt.Errorf("error statting tmpfile: %w", err)
		}

		videoInfo.FileSize = st.Size()

		return videoInfo, tmp, nil
	}

Download:
	format := video.Formats.WithAudioChannels().FindByQuality("tiny")
	videoInfo.SampleRate = 44100
	videoInfo.AudioChannels = 2

	reader, size, err := h.cl.GetStream(video, format)
	if err != nil {
		return videoInfo, nil, fmt.Errorf("error getting video stream: %w", err)
	}
	defer reader.Close()

	videoTmp, err := os.Create(tmpVideoNameAndPath)
	if err != nil {
		return videoInfo, nil, fmt.Errorf("error creating temporary video file: %w", err)
	}
	defer videoTmp.Close()
	defer os.Remove(tmpVideoNameAndPath)

	log.Logger.Infof("Size to download: %d", size)
	bar := progressbar.DefaultBytes(size, "YT Download")
	defer bar.Close()

	_, err = io.Copy(io.MultiWriter(videoTmp, bar), reader)
	if err != nil {
		return videoInfo, nil, fmt.Errorf("error copying YT stream to file: %w", err)
	}

	if h.verbose {
		if err := ffmpeg.Input(tmpVideoNameAndPath).Audio().Output(audioFile, ffmpeg.KwArgs{"sample_fmt": "s16", "ar": "44100", "ac": 2}).OverWriteOutput().WithErrorOutput(os.Stderr).Run(); err != nil {
			return videoInfo, nil, fmt.Errorf("error converting YouTubeSet download to wav: %w", err)
		}
	} else {
		if err := ffmpeg.Input(tmpVideoNameAndPath).Audio().Output(audioFile, ffmpeg.KwArgs{"sample_fmt": "s16", "ar": "44100", "ac": 2}).OverWriteOutput().Run(); err != nil {
			return videoInfo, nil, fmt.Errorf("error converting YouTubeSet download to wav: %w", err)
		}
	}

	if tmp, err = os.Open(audioFile); err != nil {
		return videoInfo, nil, fmt.Errorf("error opening WAV output file: %w", err)
	}

	st, err := tmp.Stat()
	if err != nil {
		return videoInfo, tmp, fmt.Errorf("statting WAV output file: %w", err)
	}

	videoInfo.FileSize = st.Size()

	return videoInfo, tmp, nil
}

var (
	titleCleaner, _ = regexp.Compile("[^a-zA-Z0-9 ]+")
)

func cleanTitle(title string) string {
	// Make a Regex to say we only want letters and numbers
	return strings.ReplaceAll(titleCleaner.ReplaceAllString(title, ""), " ", "_")
}

func logTrack(track, author string) {
	_, _ = pretty.Set(pretty.BgHiCyan, pretty.FgBlack, pretty.Bold).Print("🎵 Now playing")
	pretty.Unset()

	_, _ = pretty.Set(pretty.FgHiWhite, pretty.Bold).Print(" ➜ ")
	pretty.Unset()

	_, _ = pretty.Set(pretty.BgMagenta, pretty.FgWhite, pretty.Bold).Printf("%s by %s", track, author)
	pretty.Unset()
	fmt.Println()
}
