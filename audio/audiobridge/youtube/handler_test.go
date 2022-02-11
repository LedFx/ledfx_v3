package youtube

import (
	"github.com/hajimehoshi/oto"
	"ledfx/audio"
	"ledfx/audio/audiobridge"
	log "ledfx/logger"
	"math/rand"
	"testing"
	"time"
)

func TestHandlerFunctionality(t *testing.T) {
	ctx, err := oto.NewContext(44100, 2, 2, 1408)
	if err != nil {
		t.Fatalf("error starting new OTO context: %v\n", err)
	}

	wr := audio.NewAsyncMultiWriter()
	if err := wr.AddWriter(ctx.NewPlayer(), "oto"); err != nil {
		t.Fatalf("error adding oto player: %v\n", err)
	}

	h := NewHandler(&audiobridge.CallbackWrapper{
		Callback: func(buf audio.Buffer) {

		},
	}, wr, false)
	p, err := h.Play("https://www.youtube.com/watch?v=_RsiXGb1a-U")
	if err != nil {
		t.Fatalf("error playing URL: %v\n", err)
	}
	defer p.Close()

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

func TestHandlerFunctionalityPlaylist(t *testing.T) {
	ctx, err := oto.NewContext(44100, 2, 2, 1408)
	if err != nil {
		t.Fatalf("error starting new OTO context: %v\n", err)
	}

	wr := audio.NewAsyncMultiWriter()
	if err := wr.AddWriter(ctx.NewPlayer(), "oto"); err != nil {
		t.Fatalf("error adding oto player: %v\n", err)
	}

	h := NewHandler(&audiobridge.CallbackWrapper{
		Callback: func(buf audio.Buffer) {

		},
	}, wr, false)

	pp, err := h.PlayPlaylist("https://www.youtube.com/playlist?list=PLcncP1HGs_p0L1SwCfOWMjfy6vLusnJw9")
	if err != nil {
		t.Fatalf("error playing playlist: %v\n", err)
	}

	for {
		time.Sleep(5 * time.Second)
		if err := pp.PlayTrackNum(rand.Intn(pp.NumTracks())); err != nil {
			log.Logger.Errorf("Error playing next song: %v", err)
		}
	}
}
