package bridgeapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"ledfx/audio"
	"ledfx/audio/audiobridge"
	"ledfx/bridgeapi/statpoll"
	"ledfx/logger"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	ArtworkURLPath string = "/api/bridge/artwork"
)

type Server struct {
	mux        *http.ServeMux
	Br         *audiobridge.Bridge
	statPoller *statpoll.StatPoller
	upgrader   *websocket.Upgrader
}

func NewServer(callback func(buf audio.Buffer), mux *http.ServeMux) (s *Server, err error) {
	s = &Server{
		mux: mux,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			WriteBufferPool: &sync.Pool{},
		},
	}

	if s.Br, err = audiobridge.NewBridge(callback); err != nil {
		return s, fmt.Errorf("error initializing new bridge: %w", err)
	}

	s.statPoller = statpoll.New(s.Br)

	// Input setter handlers
	s.mux.HandleFunc("/api/bridge/set/input/airplay", s.handleSetInputAirPlay)
	s.mux.HandleFunc("/api/bridge/set/input/youtube", s.handleSetInputYouTube)
	s.mux.HandleFunc("/api/bridge/set/input/local", s.handleSetInputLocal)

	// Output adder handlers
	s.mux.HandleFunc("/api/bridge/add/output/airplay", s.handleAddOutputAirPlay)
	s.mux.HandleFunc("/api/bridge/add/output/local", s.handleAddOutputLocal)

	// Ctl handlers
	s.mux.HandleFunc("/api/bridge/ctl/youtube/set", s.handleCtlYouTube)
	s.mux.HandleFunc("/api/bridge/ctl/airplay/set", s.handleCtlAirPlaySet)

	// Info handlers
	s.mux.HandleFunc("/api/bridge/get/inputs/local", s.handleGetLocalInputs)

	/* TODO statpoll for these endpoints
	s.mux.HandleFunc("/api/bridge/ctl/airplay/clients", s.handleCtlAirPlayGetClients)
	s.mux.HandleFunc("/api/bridge/ctl/airplay/info", s.handleCtlAirPlayGetInfo) */

	// StatPoller handler
	s.mux.HandleFunc("/api/bridge/statpoll/ws", s.handleStatPollInitWs)

	// Artwork handler
	s.mux.HandleFunc(ArtworkURLPath, s.handleArtwork)

	return s, nil
}

func (s *Server) handleStatPollInitWs(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error upgrading connection to websocket: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errToJson(err))
		return
	}
	if err := s.statPoller.AddWebsocket(ws, r); err != nil {
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error adding websocket to statpoller: %v", err)
		_ = ws.Close()
	}
}

// ############## BEGIN AIRPLAY ##############
func (s *Server) handleSetInputAirPlay(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error reading request body: %v", err)
		w.Write(errToJson(err))
		return
	}
	logger.Logger.WithField("context", "AudioBridge").Infoln("Setting input source to AirPlay server....")
	if err := s.Br.JSONWrapper().StartAirPlayInput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error starting AirPlay input: %v", err)
		w.Write(errToJson(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (s *Server) handleAddOutputAirPlay(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error reading request body: %v", err)
		w.Write(errToJson(err))
		return
	}
	logger.Logger.WithField("context", "AudioBridge").Infoln("Adding AirPlay audio output...")
	if err := s.Br.JSONWrapper().AddAirPlayOutput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error starting AirPlay output: %v", err)
		w.Write(errToJson(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (s *Server) handleCtlAirPlaySet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("method '%s' is not allowed", r.Method)))
		return
	}

	logger.Logger.WithField("context", "AudioBridge").Infoln("Got AirPlay SET CTL request...")
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error reading request body: %v", err)
		w.Write(errToJson(err))
		return
	}

	if err := s.Br.JSONWrapper().CTL().AirPlaySet(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error getting return JSON from AirPlay CTL: %v", err)
		w.Write(errToJson(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (s *Server) handleCtlAirPlayGetClients(w http.ResponseWriter, r *http.Request) {
	logger.Logger.WithField("context", "AudioBridge").Infoln("Got AirPlay GET CTL request...")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("method '%s' is not allowed", r.Method)))
		return
	}

	clientBytes, err := s.Br.JSONWrapper().CTL().AirPlayGetClients()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error getting clients: %v", err)))
		return
	}

	w.Write(clientBytes)

}
func (s *Server) handleCtlAirPlayGetInfo(w http.ResponseWriter, r *http.Request) {
	// TODO
	w.WriteHeader(http.StatusServiceUnavailable)
}

// ############### END AIRPLAY ###############

// ############## BEGIN YOUTUBE ##############
func (s *Server) handleSetInputYouTube(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error reading request body: %v", err)
		w.Write(errToJson(err))
		return
	}
	logger.Logger.WithField("context", "AudioBridge").Infoln("Setting input source to YouTube...")
	if err := s.Br.JSONWrapper().StartYouTubeInput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error starting YouTubeSet input: %v", err)
		w.Write(errToJson(err))
		return
	}
}

func (s *Server) handleCtlYouTube(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error reading request body: %v", err)
		w.Write(errToJson(err))
		return
	}

	respBytes, err := s.Br.JSONWrapper().CTL().YouTubeSet(bodyBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error running YouTubeSet CTL action: %v", err)
		w.Write(errToJson(err))
		return
	}

	if respBytes != nil {
		w.Write(respBytes)
	}

}

func (s *Server) handleCtlYouTubeGetInfo(w http.ResponseWriter, r *http.Request) {
	ret, err := s.Br.JSONWrapper().CTL().YouTubeGetInfo()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error running YouTubeGet CTL action: %v", err)
		w.Write(errToJson(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

// ############### END YOUTUBE ###############

// ############## BEGIN LOCAL ##############
func (s *Server) handleSetInputLocal(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error reading request body: %v", err)
		w.Write(errToJson(err))
		return
	}
	logger.Logger.WithField("context", "AudioBridge").Infoln("Setting input source to local capture...")
	if err := s.Br.JSONWrapper().StartLocalInput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error starting local capture input: %v", err)
		w.Write(errToJson(err))
		return
	}
}
func (s *Server) handleAddOutputLocal(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error reading request body: %v", err)
		w.Write(errToJson(err))
		return
	}
	logger.Logger.WithField("context", "AudioBridge").Infoln("Adding local audio output...")
	if err := s.Br.JSONWrapper().AddLocalOutput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error starting local output: %v", err)
		w.Write(errToJson(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGetLocalInputs(w http.ResponseWriter, r *http.Request) {
	infos, err := audio.GetAudioDevices()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error generating input devices: %v", err)
		w.Write(errToJson(err))
		return
	}
	infoBytes, err := json.Marshal(infos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger.WithField("context", "AudioBridge").Errorf("Error marshalling input devices: %v", err)
		w.Write(errToJson(err))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(infoBytes)
}

// ############### END LOCAL ###############

// ############## BEGIN MISC ##############
func (s *Server) handleArtwork(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "image/png")
	io.Copy(w, bytes.NewReader(s.Br.Artwork()))
}

// ############### END MISC ###############

func errToJson(err error) []byte {
	b, _ := json.Marshal(map[string]string{"error": err.Error()})
	return b
}
