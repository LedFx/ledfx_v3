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
	"strings"
)

func NewServer(mux *http.ServeMux) {
	path := filepath.Join("frontend", "files")
	serveFrontend := http.FileServer(http.Dir(path))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Logger.WithField("context", "Frontend").Debugf("Serving HTTP for path: %s", r.URL.Path)
		SetContentTypeFromFilepath(r.URL.Path, w)
		serveFrontend.ServeHTTP(w, r)
	})
}

func SetContentTypeFromFilepath(fp string, w http.ResponseWriter) {
	if strings.HasSuffix(fp, "/") {
		w.Header().Set("Content-Type", "text/html")
		return
	}

	switch filepath.Ext(fp) {
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".js":
		w.Header().Set("Content-Type", "text/javascript")
	case ".map", ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpeg", ".jpg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".mp3":
		w.Header().Set("Content-Type", "audio/mpeg")
	case ".mp4":
		w.Header().Set("Content-Type", "video/mp4")
	case ".mpeg":
		w.Header().Set("Content-Type", "video/mpeg")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".wav":
		w.Header().Set("Content-Type", "audio/wav")
	case ".webm":
		w.Header().Set("Content-Type", "video/webm")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case ".xml":
		w.Header().Set("Content-Type", "application/xml")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	default:
		logger.Logger.WithField("context", "Frontend").Errorf("Could not determine content type of '%s', using default. ('text/plain')", fp)
		w.Header().Set("Content-Type", "text/plain")
	}
}

func Update() {
	logger.Logger.WithField("context", "Frontend Updater").Info("Checking for frontend updates...")
	latestFrontend := config.FrontendConfig{}
	// Check if the frontend has a new release
	resp, err := http.Get("https://api.github.com/repos/LedFx/frontend_v3/releases/latest")
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
	resp2, err := http.Get("https://github.com/LedFx/frontend_v3/releases/latest/download/frontend_v3.zip")
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
