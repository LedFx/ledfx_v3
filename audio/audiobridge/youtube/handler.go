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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	pretty "github.com/fatih/color"
	yt "github.com/kkdai/youtube/v2"
	"github.com/schollz/progressbar/v3"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.uber.org/atomic"
)

type Handler struct {
	cl         *yt.Client
	intWriter  audio.IntWriter
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

func NewHandler(intWriter audio.IntWriter, byteWriter *audio.AsyncMultiWriter, verbose bool) *Handler {
	h := &Handler{
		cl: &yt.Client{
			Debug:      false,
			HTTPClient: http.DefaultClient,
		},
		intWriter:  intWriter,
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
			intOut:  intWriter,
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

	h.p.Reset(tmp)

	return h.p, nil
}

func (h *Handler) NowPlaying() TrackInfo {
	return h.nowPlaying
}
func (h *Handler) QueuedTracks() []TrackInfo {
	return h.pp.tracks
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
	pos, err := h.p.in.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, fmt.Errorf("error getting current playback offset: %w", err)
	}
	return CompletionPercent((float32(pos) / float32(h.NowPlaying().FileSize)) * float32(100)), nil
}

func (h *Handler) PlayPlaylist(playlistURL string) (pp *PlaylistPlayer, err error) {
	pl, err := h.cl.GetPlaylist(playlistURL)
	if err != nil {
		return nil, fmt.Errorf("error getting playlist: %w", err)
	}

	h.pp.Stop()

	h.pp.tracks = make([]TrackInfo, len(pl.Videos))

	for i := range pl.Videos {
		h.pp.tracks[i] = TrackInfo{
			Artist:   pl.Videos[i].Author,
			Title:    pl.Videos[i].Title,
			Duration: SongDuration(pl.Videos[i].Duration),
			URL:      fmt.Sprintf("https://youtu.be/%s", pl.Videos[i].ID),
		}
	}
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

	// // UNIX
	// audioFile := fmt.Sprintf("/tmp/%s.wav", cleanTitle(video.Title))
	// tmpVideoName := fmt.Sprintf("%s.ytdl", util.RandString(8))
	// tmpVideoNameAndPath := filepath.Join("/tmp/", tmpVideoName)

	// Windows
	audioFile := fmt.Sprintf("tmp\\%s.wav", cleanTitle(video.Title))
	tmpVideoName := fmt.Sprintf("%s.ytdl", util.RandString(8))
	tmpVideoNameAndPath := filepath.Join("tmp", tmpVideoName)

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
	if videoInfo.SampleRate, err = strconv.ParseInt(format.AudioSampleRate, 10, 64); err != nil {
		log.Logger.WithField("category", "YT Downloader").Warnf("Error converting sample rate to integer: %v", err)
	}
	videoInfo.AudioChannels = format.AudioChannels

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
		if err := ffmpeg.Input(tmpVideoNameAndPath).Audio().Output(audioFile, ffmpeg.KwArgs{"sample_fmt": "s16", "ar": "44100"}).OverWriteOutput().WithErrorOutput(os.Stderr).Run(); err != nil {
			return videoInfo, nil, fmt.Errorf("error converting YouTubeSet download to wav: %w", err)
		}
	} else {
		if err := ffmpeg.Input(tmpVideoNameAndPath).Audio().Output(audioFile, ffmpeg.KwArgs{"sample_fmt": "s16", "ar": "44100"}).OverWriteOutput().Run(); err != nil {
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

func cleanTitle(title string) string {
	// Make a Regex to say we only want letters and numbers
	reg, _ := regexp.Compile("[^a-zA-Z0-9 ]+")
	return strings.ReplaceAll(reg.ReplaceAllString(title, ""), " ", "_")
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
