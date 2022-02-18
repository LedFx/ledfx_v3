package playback

import (
	"fmt"
	"github.com/hajimehoshi/oto/v2"
	"io"
	log "ledfx/logger"
	"ledfx/util"
)

func initCtx() error {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.WithField("category", "Local Playback Init").Errorf("Recovered during OTO ctx init: %v\n", r)
		}
	}()
	if ctx == nil {
		var err error
		var ready chan struct{}
		ctx, ready, err = oto.NewContext(44100, 2, 2)
		if err != nil {
			return fmt.Errorf("error initializing new OTO context: %w", err)
		}
		<-ready
	}
	return nil
}

var (
	ctx *oto.Context
)

type Handler struct {
	identifier string

	// pl is the player
	pl oto.Player
	pr *io.PipeReader
	pw *io.PipeWriter

	verbose bool
}

func NewHandler(verbose bool) (h *Handler, err error) {
	if err = initCtx(); err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()

	h = &Handler{
		identifier: util.RandString(8),
		pl:         ctx.NewPlayer(pr),
		pr:         pr,
		pw:         pw,
		verbose:    verbose,
	}

	h.pl.SetVolume(1)
	h.pl.Play()

	if verbose {
		log.Logger.WithField("category", "Local Playback Init").Infof("WriterID: %s", h.identifier)
	}

	return h, nil
}

func (h *Handler) Pause() {
	h.pl.Pause()
}
func (h *Handler) Resume() {
	h.pl.Play()
}

func (h *Handler) Identifier() string {
	return h.identifier
}

func (h *Handler) Quit() {
	if h.pl.IsPlaying() {
		h.pl.Close()
	}
}

func (h *Handler) Write(p []byte) (int, error) {
	return h.pw.Write(p)
}
