package playback

import (
	"fmt"
	"github.com/hajimehoshi/oto"
)

func initCtx() error {
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
	// pl is the player
	pl *oto.Player
}

func NewHandler() (h *Handler, err error) {
	if err = initCtx(); err != nil {
		return nil, err
	}

	h = &Handler{
		pl: ctx.NewPlayer(),
	}

	return h, nil
}

func (h *Handler) Quit() {
	h.pl.Close()
}

func (h *Handler) Write(p []byte) (int, error) {
	return h.pl.Write(p)
}
