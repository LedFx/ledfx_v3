package playback

import (
	"fmt"
	"github.com/hajimehoshi/oto"
	"math/rand"
	"strconv"
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
	identifier string
	// pl is the player
	pl *oto.Player
}

func NewHandler() (h *Handler, err error) {
	if err = initCtx(); err != nil {
		return nil, err
	}

	randStr := func() string {
		b := make([]byte, 16)
		for i := 0; i < 16; i++ {
			b = append(b, []byte(strconv.Itoa(rand.Intn(9)))...)
		}
		return string(b)
	}()

	h = &Handler{
		pl:         ctx.NewPlayer(),
		identifier: randStr,
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
