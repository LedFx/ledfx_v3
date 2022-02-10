package youtube

import (
	"github.com/hajimehoshi/oto"
	"ledfx/audio"
	"ledfx/audio/audiobridge"
	"testing"
	"time"
)

func TestHandlerFunctionality(t *testing.T) {
	ctx, err := oto.NewContext(44100, 2, 2, 1408)
	if err != nil {
		t.Fatalf("error starting new OTO context: %v\n", err)
	}
	defer ctx.Close()

	wr := audio.NewAsyncMultiWriter()
	if err := wr.AddWriter(ctx.NewPlayer(), "oto"); err != nil {
		t.Fatalf("error adding oto player: %v\n", err)
	}

	h := NewHandler(&audiobridge.CallbackWrapper{
		Callback: func(buf audio.Buffer) {

		},
	}, wr, false)
	p, err := h.Play("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("error rickrolling: %v\n", err)
	}
	defer p.Stop()
	go func() {
		for {
			time.Sleep(5 * time.Second)
			p.Pause()
			time.Sleep(2 * time.Second)
			p.Unpause()
		}
	}()
	t.Log("playing...")
	if err := p.Start(); err != nil {
		t.Fatalf("error starting player: %v\n", err)
	}
}
