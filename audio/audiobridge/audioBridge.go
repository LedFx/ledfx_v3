package audiobridge

import (
	"io"
)

// NewBridge initializes a new bridge between a source and destination audio device.
func NewBridge(ledFxWriter io.Writer) (br *Bridge, err error) {
	br = &Bridge{
		ledFxWriter: ledFxWriter,
		done:        make(chan bool),
		inputType:   inputType(-1), // -1 signifies undefined
	}
	return br, nil
}

func (br *Bridge) Wait() {
	<-br.done
}

func (br *Bridge) stop(notifyDone bool) {
	if notifyDone {
		defer func() {
			go func() {
				br.done <- true
			}()
		}()
	}

	if br.airplay != nil {
		br.airplay.Stop()
	}

	if br.local != nil {
		br.local.Stop()
	}

}

// Stop stops the bridge. Any further references to 'br *Bridge'
// may cause a runtime panic.
func (br *Bridge) Stop() {
	if br != nil {
		br.stop(true)
	}
}
