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

func NewServer(mux *http.ServeMux) {
	serveFrontend := http.FileServer(http.Dir("frontend/files"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Logger.WithField("context", "Frontend").Debugf("Serving HTTP for path: %s", r.URL.Path)
		serveFrontend.ServeHTTP(w, r)
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
		logger.Logger.WithField("context", "Frontend Updater").Infof("New frontend available! Updating to %s", latestFrontend.TagName)
	}

	// If an update is available, download it
	resp2, err := http.Get("https://github.com/YeonV/LedFx-Frontend-v2/releases/latest/download/ledfx_frontend_v2.zip")
	if err != nil || resp2.StatusCode != 200 {
		logger.Logger.WithField("context", "Frontend Updater").Error(err)
		return
	}
	defer resp2.Body.Close()

	// get absolute location of executable, where we will install frontend files
	ex, err := util.ExecDir()
	if err != nil {
		logger.Logger.WithField("context", "Frontend Updater").Error(err)
		return
	} else {
		logger.Logger.WithField("context", "Frontend Updater").Debugf("Extracting to %s", ex)
	}

	// Create the new file
	zipPath := filepath.Join(ex, "new_frontend.zip")
	out, err := os.Create(zipPath)
	defer os.RemoveAll(zipPath)
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
	filesPath := filepath.Join(ex, "frontend", "files")
	os.MkdirAll(filesPath, os.ModePerm)
	if _, err := os.Stat(filesPath); err == nil {
		os.RemoveAll(filesPath)
		logger.Logger.WithField("context", "Frontend Updater").Debug("Deleted old frontend")
	}

	// Extract frontend
	err = util.Unzip(zipPath, filesPath)
	if err != nil {
		logger.Logger.WithField("context", "Frontend Updater").Error(err)
		return
	}

	// Save this version to config
	config.SetFrontend(latestFrontend)
	logger.Logger.WithField("context", "Frontend Updater").Info("Got latest Frontend")
}
