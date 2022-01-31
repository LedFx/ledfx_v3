package audiobridge

import (
	"fmt"
	"github.com/hajimehoshi/oto"
	"ledfx/audio/audiobridge/loopback"
)

func (br *Bridge) StartLocalInput() (err error) {
	if br.inputType != -1 {
		return fmt.Errorf("an input source has already been defined for this bridge")
	}

	br.inputType = inputTypeLocal

	if br.local == nil {
		br.local = new(LocalHandler)
	}

	if br.local.loopback == nil {
		if br.local.loopback, err = loopback.New(br.ledFxWriter); err != nil {
			return fmt.Errorf("error initializing new loopback device: %w", err)
		}
	}

	return nil
}

func (br *Bridge) AddLocalOutput() (err error) {
	if br.local == nil {
		br.local = new(LocalHandler)
	}
	if br.local.ctx == nil {
		if br.local.ctx, err = oto.NewContext(44100, 2, 2, 12000); err != nil {
			return fmt.Errorf("error creating new OTO context: %w", err)
		}
	}

	if br.local.player == nil {
		br.local.player = br.local.ctx.NewPlayer()
	}

	if err = br.wireLocalOutput(br.local.player); err != nil {
		return fmt.Errorf("error wiring local output: %w", err)
	}

	return nil
}

type LocalHandler struct {
	ctx      *oto.Context
	player   *oto.Player
	loopback *loopback.Loopback
}

func (lh *LocalHandler) Stop() {
	if lh.ctx != nil {
		_ = lh.ctx.Close()
	}
}
