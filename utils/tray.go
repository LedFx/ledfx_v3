//go:generate goversioninfo -icon=assets/logo.ico
package utils

import (
	_ "embed"

	"github.com/getlantern/systray"
)

//go:embed assets/logo.ico
var logo []byte

func OnReady() {
	systray.SetIcon(logo)
	systray.SetTooltip("LedFx-Go")
	mOpen := systray.AddMenuItem("Open", "Open LedFx in Browser")
	mGithub := systray.AddMenuItem("Github", "Open LedFx in Browser")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	for {
		select {
		case <-mOpen.ClickedCh:
			Openbrowser("http://localhost:8080")
		case <-mGithub.ClickedCh:
			Openbrowser("https://github.com/YeonV/ledfx-go")
		case <-mQuit.ClickedCh:
			systray.Quit()
		}
	}
}
