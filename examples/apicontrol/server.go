package main

import (
	"fmt"
	"io/ioutil"
	"ledfx/audio"
	"ledfx/audio/audiobridge"
	log "ledfx/logger"
	"net/http"
)

type Server struct {
	br *audiobridge.Bridge
	sv *http.Server
}

func NewServer(callback func(buf audio.Buffer)) (s *Server, err error) {
	s = &Server{
		sv: &http.Server{
			Addr: "0.0.0.0:8080",
		},
	}
	if s.br, err = audiobridge.NewBridge(callback); err != nil {
		return nil, fmt.Errorf("error initializing new bridge: %w", err)
	}

	// Input setter handlers
	http.HandleFunc("/set/input/airplay", s.handleSetInputAirPlay)
	http.HandleFunc("/set/input/youtube", s.handleSetInputYouTube)
	http.HandleFunc("/set/input/capture", s.handleSetInputCapture)

	// Output adder handlers
	http.HandleFunc("/add/output/airplay", s.handleAddOutputAirPlay)
	http.HandleFunc("/add/output/local", s.handleAddOutputLocal)

	// Ctl handlers
	http.HandleFunc("/ctl/youtube", s.handleCtlYouTube)
	http.HandleFunc("/ctl/airplay/set", s.handleCtlAirPlaySet)
	http.HandleFunc("/ctl/airplay/clients", s.handleCtlAirPlayGetClients)
	http.HandleFunc("/ctl/airplay/info", s.handleCtlAirPlayGetInfo)
	return s, nil
}

func (s *Server) Serve(ip string, port int) error {
	s.sv.Addr = fmt.Sprintf("%s:%d", ip, port)
	log.Logger.Warnf("Serving on %s", s.sv.Addr)
	return s.sv.ListenAndServe()
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
	log.Logger.Infoln("Setting input source to YouTube....")
	if err := s.br.JSONWrapper().StartYouTubeInput(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error starting YouTube input: %v", err)
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
	log.Logger.Infof("Got YouTube CTL request...")
	if err := s.br.JSONWrapper().CTL().YouTube(bodyBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Logger.Errorf("Error running YouTube CTL action: %v", err)
		w.Write(errToBytes(err))
		return
	}
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
