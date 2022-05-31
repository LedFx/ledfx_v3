//go:generate goversioninfo -icon=assets/logo.ico
package unstructuredold

import (
	_ "embed"
	log "ledfx/logger"

	"fyne.io/systray"
)

//go:embed assets/logo.ico
var logo []byte

func OnReady() {
	systray.SetIcon(logo)
	systray.SetTooltip("LedFx-Go")
	mOpen := systray.AddMenuItem("Open", "Open LedFx Frontend in Browser")
	mGithub := systray.AddMenuItem("Github", "Open LedFx Github in Browser")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Shutdown LedFx")
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				Openbrowser("http://localhost:8080/#/?newCore=1")
			case <-mGithub.ClickedCh:
				Openbrowser("https://github.com/LedFx/ledfx_rewrite")
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func OnExit() {
	log.Logger.WithField("category", "Systray Handler").Warnln("Closing systray...")
}
