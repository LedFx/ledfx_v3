package playback

import (
	"fmt"
	"github.com/carterpeel/oto/v2"
	"io"
)

func initCtx() error {
	if ctx == nil {
		var ready chan struct{}
		var err error
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
	// pl is the player
	pl oto.Player

	// pr is the pipe reader
	pr *io.PipeReader

	// pw is the pipe writer
	pw *io.PipeWriter
}

func NewHandler() (h *Handler, err error) {
	if err = initCtx(); err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()

	pl := ctx.NewPlayer(pr)

	h = &Handler{
		pr: pr,
		pw: pw,
		pl: pl,
	}

	h.pl.Play(false)

	return h, nil
}

func (h *Handler) Quit() {
	h.pr.CloseWithError(io.EOF)
	h.pw.CloseWithError(io.EOF)
	h.pl.Close()
	h.pl = nil
}

func (h *Handler) Write(b []byte) (n int, err error) {
	return h.pw.Write(b)
}
