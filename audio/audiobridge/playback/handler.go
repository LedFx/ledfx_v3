package playback

import (
	"fmt"
	"github.com/hajimehoshi/oto"
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
		ctx, err = oto.NewContext(44100, 2, 2, 1408)
		if err != nil {
			return fmt.Errorf("error initializing new OTO context: %w", err)
		}
	}
	return nil
}

var (
	ctx *oto.Context
)

type Handler struct {
	identifier string
	// pl is the player
	pl      *oto.Player
	verbose bool
}

func NewHandler(verbose bool) (h *Handler, err error) {
	if err = initCtx(); err != nil {
		return nil, err
	}

	h = &Handler{
		pl:         ctx.NewPlayer(),
		identifier: util.RandString(8),
		verbose:    verbose,
	}

	if verbose {
		log.Logger.WithField("category", "Local Playback Init").Infof("Identifier: %s\n", h.identifier)
	}

	return h, nil
}

func (h *Handler) Identifier() string {
	return h.identifier
}

func (h *Handler) Quit() {
	h.pl.Close()
}

func (h *Handler) Write(p []byte) (int, error) {
	return h.pl.Write(p)
}
