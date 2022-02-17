package bridgeapi

import (
	"fmt"
	"io/ioutil"
	"ledfx/audio"
	"ledfx/audio/audiobridge"
	log "ledfx/logger"
	"net/http"
)

type Server struct {
	mux *http.ServeMux
	br  *audiobridge.Bridge
}

func NewServer(callback func(buf audio.Buffer), mux *http.ServeMux) (err error) {
	s := &Server{
		mux: mux,
	}
	if s.br, err = audiobridge.NewBridge(callback); err != nil {
		return fmt.Errorf("error initializing new bridge: %w", err)
	}

	// Input setter handlers
	s.mux.HandleFunc("/api/bridge/set/input/airplay", s.handleSetInputAirPlay)
	s.mux.HandleFunc("/api/bridge/set/input/youtube", s.handleSetInputYouTube)
	s.mux.HandleFunc("/api/bridge/set/input/capture", s.handleSetInputCapture)

	// Output adder handlers
	s.mux.HandleFunc("/api/bridge/add/output/airplay", s.handleAddOutputAirPlay)
	s.mux.HandleFunc("/api/bridge/add/output/local", s.handleAddOutputLocal)

	// Ctl handlers
	s.mux.HandleFunc("/api/bridge/ctl/youtube/set", s.handleCtlYouTube)
	s.mux.HandleFunc("/api/bridge/ctl/youtube/info", s.handleCtlYouTubeGetInfo)

	s.mux.HandleFunc("/api/bridge/ctl/airplay/set", s.handleCtlAirPlaySet)
	s.mux.HandleFunc("/api/bridge/ctl/airplay/clients", s.handleCtlAirPlayGetClients)
	s.mux.HandleFunc("/api/bridge/ctl/airplay/info", s.handleCtlAirPlayGetInfo)
	return nil
}

// ############## BEGIN AIRPLAY ##############
func (s *Server) handleSetInputAirPlay(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	log.Logger.Infoln("Setting input source to AirPlay server....")
	if err := s.br.JSONWrapper().StartAirPlayInput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error starting AirPlay input: %v", err)
		w.Write(errToBytes(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (s *Server) handleAddOutputAirPlay(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	log.Logger.Infoln("Adding AirPlay audio output...")
	if err := s.br.JSONWrapper().AddAirPlayOutput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error starting AirPlay output: %v", err)
		w.Write(errToBytes(err))
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

	log.Logger.Infoln("Got AirPlay SET CTL request...")
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}

	if err := s.br.JSONWrapper().CTL().AirPlaySet(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error getting return JSON from AirPlay CTL: %v", err)
		w.Write(errToBytes(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (s *Server) handleCtlAirPlayGetClients(w http.ResponseWriter, r *http.Request) {
	log.Logger.Infoln("Got AirPlay GET CTL request...")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("method '%s' is not allowed", r.Method)))
		return
	}

	clientBytes, err := s.br.JSONWrapper().CTL().AirPlayGetClients()
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
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	log.Logger.Infoln("Setting input source to YouTubeSet....")
	if err := s.br.JSONWrapper().StartYouTubeInput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error starting YouTubeSet input: %v", err)
		w.Write(errToBytes(err))
		return
	}
}

func (s *Server) handleCtlYouTube(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	if err := s.br.JSONWrapper().CTL().YouTubeSet(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error running YouTubeSet CTL action: %v", err)
		w.Write(errToBytes(err))
		return
	}
}

func (s *Server) handleCtlYouTubeGetInfo(w http.ResponseWriter, r *http.Request) {
	log.Logger.Infoln("Got YouTubeSet GET CTL request...")
	ret, err := s.br.JSONWrapper().CTL().YouTubeGetInfo()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error running YouTubeGet CTL action: %v", err)
		w.Write(errToBytes(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

// ############### END YOUTUBE ###############

// ############## BEGIN LOCAL ##############
func (s *Server) handleSetInputCapture(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	log.Logger.Infoln("Setting input source to local capture...")
	if err := s.br.JSONWrapper().StartLocalInput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error starting local capture input: %v", err)
		w.Write(errToBytes(err))
		return
	}
}
func (s *Server) handleAddOutputLocal(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error reading request body: %v", err)
		w.Write(errToBytes(err))
		return
	}
	log.Logger.Infoln("Adding local audio output...")
	if err := s.br.JSONWrapper().AddLocalOutput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error starting local output: %v", err)
		w.Write(errToBytes(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ############### END LOCAL ###############

func errToBytes(err error) []byte {
	return []byte(err.Error() + "\n")
}
