package youtube

import (
	"fmt"
	"github.com/dustin/go-humanize"
	pretty "github.com/fatih/color"
	yt "github.com/kkdai/youtube/v2"
	"github.com/schollz/progressbar/v3"
	ffmpeg "github.com/u2takey/ffmpeg-go"
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
)

type Handler struct {
	cl         *yt.Client
	intWriter  audio.IntWriter
	byteWriter *audio.AsyncMultiWriter
	verbose    bool
	p          *Player
	pp         *PlaylistPlayer
}

func (h *Handler) Quit() {
	if h.p != nil {
		h.p.Close()
	}
	if h.pp != nil {
		h.pp.Stop()
	}
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
		p: &Player{
			mu:      &sync.Mutex{},
			paused:  false,
			unpause: make(chan bool),
			done:    false,
			in:      nil,
			out:     byteWriter,
			intOut:  intWriter,
		},
	}
	h.pp = &PlaylistPlayer{
		h:        h,
		trackNum: -1,
		tracks:   nil,
	}
	return h
}

func (h *Handler) Play(url string) (p *Player, err error) {
	videoInfo, tmp, err := h.downloadToMP3(url)
	if err != nil {
		return nil, fmt.Errorf("error writing video data to tmpfile: %w", err)
	}

	if h.verbose {
		log.Logger.WithField("category", "YT Player").Infof(
			"[CREATOR=\"%s\", SAMPLERATE=\"%d\", SIZE=\"%s\", CHANNELS=\"%d\", DURATION=\"%v\"]",
			videoInfo.Creator,
			videoInfo.SampleRate,
			humanize.Bytes(uint64(videoInfo.WavSize)),
			videoInfo.Channels,
			videoInfo.duration,
		)
	}

	_, _ = pretty.Set(pretty.BgHiCyan, pretty.FgBlack, pretty.Bold).Print("🎵 Now playing")
	pretty.Unset()

	_, _ = pretty.Set(pretty.FgHiWhite, pretty.Bold).Print(" ➜ ")
	pretty.Unset()

	_, _ = pretty.Set(pretty.BgMagenta, pretty.FgWhite, pretty.Bold).Printf("%s by %s", videoInfo.Title, videoInfo.Creator)
	pretty.Unset()
	fmt.Println()

	h.p.Reset(tmp)

	return h.p, nil
}

func (h *Handler) PlayPlaylist(playlistURL string) (pp *PlaylistPlayer, err error) {
	pl, err := h.cl.GetPlaylist(playlistURL)
	if err != nil {
		return nil, fmt.Errorf("error getting playlist: %w", err)
	}

	h.pp.Stop()

	h.pp.tracks = make([]string, len(pl.Videos))

	for i := range pl.Videos {
		h.pp.tracks[i] = fmt.Sprintf("https://youtu.be/%s", pl.Videos[i].ID)
	}
	return h.pp, nil
}

type VideoInfo struct {
	Title      string
	Creator    string
	WavSize    int64
	SampleRate int
	Channels   int
	duration   time.Duration
}

func (h *Handler) downloadToMP3(url string) (videoInfo VideoInfo, tmp *os.File, err error) {
	video, err := h.cl.GetVideo(url)
	if err != nil {
		return videoInfo, nil, fmt.Errorf("error getting video: %w", err)
	}

	videoInfo = VideoInfo{
		Title:      video.Title,
		Creator:    video.Author,
		duration:   video.Duration,
		SampleRate: -1,
	}

	audioFile := fmt.Sprintf("/tmp/%s.wav", cleanTitle(video.Title))
	tmpVideoName := fmt.Sprintf("%s.ytdl", util.RandString(8))
	tmpVideoNameAndPath := filepath.Join("/tmp/", tmpVideoName)

	if util.FileExists(audioFile) {
		if tmp, err = os.Open(audioFile); err != nil {
			goto Download
		}
		st, err := tmp.Stat()
		if err != nil {
			return videoInfo, nil, fmt.Errorf("error statting tmpfile: %w", err)
		}

		videoInfo.WavSize = st.Size()

		return videoInfo, tmp, nil
	}

Download:
	format := video.Formats.WithAudioChannels().FindByQuality("tiny")
	if videoInfo.SampleRate, err = strconv.Atoi(format.AudioSampleRate); err != nil {
		log.Logger.WithField("category", "YT Downloader").Warnf("Error converting sample rate to integer: %v", err)
	}
	videoInfo.Channels = format.AudioChannels

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
			return videoInfo, nil, fmt.Errorf("error converting YouTube download to wav: %w", err)
		}
	} else {
		if err := ffmpeg.Input(tmpVideoNameAndPath).Audio().Output(audioFile, ffmpeg.KwArgs{"sample_fmt": "s16", "ar": "44100"}).OverWriteOutput().Run(); err != nil {
			return videoInfo, nil, fmt.Errorf("error converting YouTube download to wav: %w", err)
		}
	}

	if tmp, err = os.Open(audioFile); err != nil {
		return videoInfo, nil, fmt.Errorf("error opening WAV output file: %w", err)
	}

	st, err := tmp.Stat()
	if err != nil {
		return videoInfo, tmp, fmt.Errorf("statting WAV output file: %w", err)
	}

	videoInfo.WavSize = st.Size()

	return videoInfo, tmp, nil
}

func cleanTitle(title string) string {
	// Make a Regex to say we only want letters and numbers
	reg, _ := regexp.Compile("[^a-zA-Z0-9 ]+")
	return strings.ReplaceAll(reg.ReplaceAllString(title, ""), " ", "_")
}
