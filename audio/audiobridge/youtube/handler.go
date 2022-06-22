package youtube

import (
	"errors"
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

	ffmpeg "github.com/carterpeel/ffmpeg-go"
	pretty "github.com/fatih/color"
	yt "github.com/kkdai/youtube/v2"
	progressbar "github.com/schollz/progressbar/v3"
	"go.uber.org/atomic"
)

type Handler struct {
	cl         *yt.Client
	byteWriter *audio.AsyncMultiWriter
	p          *Player

	stopped bool

	nowPlaying TrackInfo
	history    []TrackInfo
}

func (h *Handler) Quit() {
	h.stopped = true
	h.p.Close()

}
func (h *Handler) Stopped() bool {
	return h.stopped
}

func (h *Handler) Player() *Player {
	return h.p
}

func NewHandler(byteWriter *audio.AsyncMultiWriter) *Handler {
	h := &Handler{
		cl: &yt.Client{
			Debug:      false,
			HTTPClient: http.DefaultClient,
		},
		byteWriter: byteWriter,
		history:    make([]TrackInfo, 0),
		p: &Player{
			mu:              &sync.Mutex{},
			isDone:          atomic.NewBool(false),
			done:            make(chan struct{}),
			paused:          atomic.NewBool(false),
			pause:           make(chan struct{}),
			unpause:         make(chan struct{}),
			playing:         atomic.NewBool(false),
			play:            make(chan []byte),
			cycling:         atomic.NewBool(false),
			next:            make(chan struct{}),
			nextDone:        make(chan struct{}),
			prev:            make(chan struct{}),
			prevDone:        make(chan struct{}),
			playByIndex:     make(chan int32),
			playByIndexDone: make(chan struct{}),
			out:             byteWriter,
			elapsed:         atomic.NewDuration(0),
			trackMu:         &sync.Mutex{},
			trackNum:        atomic.NewInt32(0),
			tracks:          make([]TrackInfo, 0),
			trackPaths:      make([]string, 0),
		},
	}
	h.p.h = h

	return h
}

func (h *Handler) downloadWAV(info TrackInfo, current, max int, clearBar bool) (path string, err error) {
	if info.video == nil {
		if info.video, err = h.cl.GetVideo(info.URL); err != nil {
			return path, fmt.Errorf("error getting video metadata: %w", err)
		}
	}

	path = filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s.wav", cleanString(info.Title), cleanString(info.Artist)))

	// Don't re-download files we already have
	if util.FileExists(path) {
		log.Logger.WithField("context", "YouTube WAV Downloader").Infof("Found cached audio file: %q", filepath.Base(path))
		return path, nil
	}

	tmpVideoName := fmt.Sprintf("%s.ytdl", util.RandString(8))
	tmpVideoNameAndPath := filepath.Join(os.TempDir(), tmpVideoName)

	withAudioChannels := info.video.Formats.WithAudioChannels()
	withAudioChannels.Sort()
	if len(withAudioChannels) == 0 {
		return path, errors.New("no audio channels found in video")
	}

	var was403 bool
	var format *yt.Format
Retry:
	if was403 {
		format = &withAudioChannels[0]
	} else {
		var curQualityValue uint8
		var curAudioCh int
		var curAudioSampleRate int

		for i := range withAudioChannels {
			curFmt := &withAudioChannels[i]
			quality, err := audioQualityToInt(curFmt.AudioQuality)
			if err != nil {
				log.Logger.Errorf("error: %v", err)
				continue
			}

			sampleRate, _ := strconv.Atoi(curFmt.AudioSampleRate)

			switch {
			case curFmt.ItagNo == 137: // ItagNo 137 downloads very slowly
				continue
			case quality > curQualityValue:
				fallthrough
			case quality >= curQualityValue && curFmt.AudioChannels >= curAudioCh:
				fallthrough
			case quality >= curQualityValue && curFmt.AudioChannels >= curAudioCh && sampleRate >= curAudioSampleRate:
				curQualityValue = quality
				curAudioCh = curFmt.AudioChannels
				curAudioSampleRate = sampleRate
				format = curFmt
			}
		}
	}

	if format == nil {
		return path, errors.New("could not find suitable audio stream in video")
	}

	reader, size, err := h.cl.GetStream(info.video, format)
	if err != nil {
		return path, fmt.Errorf("error getting video stream: %w", err)
	}
	defer reader.Close()

	videoTmp, err := os.Create(tmpVideoNameAndPath)
	if err != nil {
		return path, fmt.Errorf("error creating temporary video file: %w", err)
	}
	defer videoTmp.Close()
	defer os.Remove(tmpVideoNameAndPath)

	opts := []progressbar.Option{
		progressbar.OptionFullWidth(),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetDescription(fmt.Sprintf("[_black_][bold][cyan][%d/%d] [yellow]Downloading: [bold][red]%q[reset][_black_][cyan]", current, max, strings.TrimSuffix(info.video.Title, " "))),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[_black_][bold][light_green]▰[reset]",
			SaucerHead:    "[_black_][bold][light_green]█[reset]",
			SaucerPadding: "[_black_]▱",
			BarStart:      "[bold][_black_][cyan]{",
			BarEnd:        "[bold][_black_][cyan]}[reset]",
		}),
	}

	if clearBar == true {
		opts = append(opts, progressbar.OptionClearOnFinish())
	}
	bar := progressbar.NewOptions64(size, opts...)
	defer bar.Close()

	_, err = io.Copy(io.MultiWriter(videoTmp, bar), reader)
	if err != nil {
		if strings.HasSuffix(err.Error(), "403") && !was403 {
			_ = bar.Clear()
			was403 = true
			goto Retry
		}
		return path, fmt.Errorf("error copying YT stream to file: %w", err)
	}

	if err := ffmpeg.Input(tmpVideoNameAndPath).Audio().Output(path, ffmpeg.KwArgs{"sample_fmt": "s16", "ar": "44100", "ac": 2}).OverWriteOutput().Run(); err != nil {
		return path, fmt.Errorf("error converting YouTubeSet download to wav: %w", err)
	}

	return path, nil
}

var (
	stringCleaner, _ = regexp.Compile("[^a-zA-Z0-9 ]+")
)

func cleanString(title string) string {
	// Make a Regex to say we only want letters and numbers
	return strings.ReplaceAll(stringCleaner.ReplaceAllString(title, ""), " ", "_")
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

const (
	aqHigh   = "AUDIO_QUALITY_HIGH"
	aqMedium = "AUDIO_QUALITY_MEDIUM"
	aqLow    = "AUDIO_QUALITY_LOW"
)

func audioQualityToInt(qualityStr string) (uint8, error) {
	switch qualityStr {
	case aqHigh:
		return 2, nil
	case aqMedium:
		return 1, nil
	case aqLow:
		return 0, nil
	default:
		return 0, fmt.Errorf("unknown audio quality value %q", qualityStr)
	}
}
