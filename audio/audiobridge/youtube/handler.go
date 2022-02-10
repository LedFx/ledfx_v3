package youtube

import (
	"fmt"
	yt "github.com/kkdai/youtube/v2"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"ledfx/audio"
	"ledfx/util"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Handler struct {
	cl         *yt.Client
	intWriter  audio.IntWriter
	byteWriter *audio.AsyncMultiWriter
	verbose    bool
	p          *Player
}

func NewHandler(intWriter audio.IntWriter, byteWriter *audio.AsyncMultiWriter, verbose bool) *Handler {
	return &Handler{
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
}

func (h *Handler) Play(url string) (p *Player, err error) {
	tmp, err := h.downloadToMP3(url)
	if err != nil {
		return nil, fmt.Errorf("error writing video data to tmpfile: %w", err)
	}

	h.p.Reset(tmp)

	return h.p, nil
}

func (h *Handler) downloadToMP3(url string) (*os.File, error) {
	video, err := h.cl.GetVideo(url)
	if err != nil {
		return nil, fmt.Errorf("error getting video: %w", err)
	}
	reader, _, err := h.cl.GetStream(video, video.Formats.FindByQuality("medium"))
	if err != nil {
		return nil, fmt.Errorf("error getting video stream: %w", err)
	}
	defer reader.Close()

	tmpVideoName := fmt.Sprintf("%s.ytdl", util.RandString(8))
	tmpVideoNameAndPath := filepath.Join("/tmp/", tmpVideoName)

	videoTmp, err := os.Create(tmpVideoNameAndPath)
	if err != nil {
		return nil, fmt.Errorf("error creating temporary video file: %w", err)
	}
	defer videoTmp.Close()
	defer os.Remove(tmpVideoNameAndPath)

	if _, err := io.Copy(videoTmp, reader); err != nil {
		return nil, fmt.Errorf("error copying YT stream to file: %w", err)
	}

	audioFile := fmt.Sprintf("/tmp/%s.wav", strings.ReplaceAll(video.Title, " ", "_"))

	if h.verbose {
		if err := ffmpeg.Input(tmpVideoNameAndPath).Audio().Output(audioFile, ffmpeg.KwArgs{"sample_fmt": "s16", "ar": "44100"}).OverWriteOutput().WithErrorOutput(os.Stderr).Run(); err != nil {
			return nil, fmt.Errorf("error converting YouTube download to wav: %w", err)
		}
	} else {
		if err := ffmpeg.Input(tmpVideoNameAndPath).Audio().Output(audioFile, ffmpeg.KwArgs{"sample_fmt": "s16", "ar": "44100"}).OverWriteOutput().Run(); err != nil {
			return nil, fmt.Errorf("error converting YouTube download to wav: %w", err)
		}
	}

	tmp, err := os.Open(audioFile)
	if err != nil {
		return nil, fmt.Errorf("error opening WAV output file: %w", err)
	}

	return tmp, nil
}
