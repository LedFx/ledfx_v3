package frontend

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"ledfx/config"
	"ledfx/logger"
	"ledfx/util"
	"net/http"
	"os"
	"path/filepath"
)

func ServeHttp() {
	Update()
	serveFrontend := http.FileServer(http.Dir("frontend"))
	// api.HandleApi()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Logger.WithField("context", "Frontend").Debugf("Request asked for %s", r.URL.Path)
		if filepath.Ext(r.URL.Path) == "" {
			logger.Logger.WithField("context", "Frontend").Debugln("Serving index.html")
			http.ServeFile(w, r, "frontend/index.html")
		} else {
			logger.Logger.WithField("context", "Frontend").Debugf("Serving HTTP for path: %s", r.URL.Path)
			serveFrontend.ServeHTTP(w, r)
		}
	})
}

func Update() {
	logger.Logger.WithField("context", "Frontend Updater").Info("Checking for frontend updates...")
	latestFrontend := config.FrontendConfig{}
	// Check if the frontend has a new release
	resp, err := http.Get("https://api.github.com/repos/YeonV/LedFx-Frontend-v2/releases/latest")
	if err != nil || resp.StatusCode != 200 {
		logger.Logger.WithField("context", "Frontend Updater").Error(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Logger.WithField("context", "Frontend Updater").Error(err)
		return
	}
	json.Unmarshal(body, &latestFrontend)

	// Compare latest release commit to saved commit
	if latestFrontend.Commit == config.GetFrontend().Commit {
		logger.Logger.WithField("context", "Frontend Updater").Debug("Frontend is up to date")
		return
	} else {
		logger.Logger.WithField("context", "Frontend Updater").Info("New frontend available. Updating...")
	}

	// If an update is available, download it
	resp2, err := http.Get("https://github.com/YeonV/LedFx-Frontend-v2/releases/latest/download/ledfx_frontend_v2.zip")
	if err != nil || resp2.StatusCode != 200 {
		logger.Logger.WithField("context", "Frontend Updater").Error(err)
		return
	}
	defer resp2.Body.Close()

	// Create the new file
	out, err := os.Create("new_frontend.zip")
	defer os.RemoveAll("new_frontend.zip")
	defer out.Close()
	if err != nil {
		logger.Logger.WithField("context", "Frontend Updater").Error(err)
		return
	}

	// Write the body to file
	_, err = io.Copy(out, resp2.Body)
	if err != nil {
		logger.Logger.WithField("context", "Frontend Updater").Error(err)
		return
	}

	// Delete old files
	if _, err := os.Stat("frontend/files"); err == nil {
		os.RemoveAll("frontend/files")
		logger.Logger.WithField("context", "Frontend Updater").Debug("Deleted old frontend")
	}

	// Extract frontend
	util.Unzip("new_frontend.zip", "frontend/files")
	defer os.RemoveAll("new_frontend.zip")

	// Save this version to config
	config.SetFrontend(latestFrontend)
	logger.Logger.WithField("context", "Frontend Updater").Info("Got latest Frontend")
}
