package main

import (
	_ "embed"
	"fmt"
	"ledfx/event"
	"ledfx/logger"
	"ledfx/util"
	"os"

	"fyne.io/systray"
)

func StartTray(url string) func() {
	return func() {
		systray.SetIcon(util.Logo)
		systray.SetTooltip("LedFx")
		mOpen := systray.AddMenuItem("Open", "Open LedFx Web Interface in Browser")
		mGithub := systray.AddMenuItem("Github", "Open LedFx Github in Browser")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Shutdown LedFx")
		go func() {
			for {
				select {
				case <-mOpen.ClickedCh:
					util.OpenBrowser(fmt.Sprintf("http://%s/#/?newCore=1", url))
				case <-mGithub.ClickedCh:
					util.OpenBrowser("https://github.com/LedFx/ledfx_rewrite")
				case <-mQuit.ClickedCh:
					event.Invoke(event.Shutdown, map[string]interface{}{})
					return
				}
			}
		}()
	}
}

func StopTray() {
	// TODO kill ledfx from here. need to emit a broadcast event.
	logger.Logger.WithField("context", "Systray Handler").Warnln("Removed systray icon")
	os.Exit(0)
}
