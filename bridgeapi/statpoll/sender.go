package statpoll

import (
	"github.com/gorilla/websocket"
	"ledfx/audio/audiobridge"
	"ledfx/audio/audiobridge/youtube"
	"ledfx/integrations/airplay2"
	log "ledfx/logger"
	"time"
)

type Response struct {
	Type      ReqType `json:"type"`
	Iteration int     `json:"iteration"`
	Value     interface{}
}

// valueBridgeInfo contains all information on audiobridge.Bridge
type valueBridgeInfo struct {
	InputType string                    `json:"input_type"`
	Outputs   []*audiobridge.OutputInfo `json:"outputs"`
}

func (s *StatPoller) sendBridgeInfo(n int, interval time.Duration, ws *websocket.Conn) {
	if s.sendingBridgeInfo.IsSet() {
		s.stopBridgeInfo()
	}

	s.bridgeDispatchMu.Lock()
	defer s.bridgeDispatchMu.Unlock()

	s.sendingBridgeInfo.Set()

	tick := time.NewTicker(interval)
	defer tick.Stop()

Top:
	for i := 0; i != n; i++ {
		select {
		case <-tick.C:
			if err := ws.WriteJSON(&Response{
				Type:      RqtBridgeInfo,
				Iteration: i,
				Value: &valueBridgeInfo{
					InputType: s.br.Info().InputType(),
					Outputs:   s.br.Info().AllOutputs(),
				},
			}); err != nil {
				log.Logger.WithField("category", "StatPoll BridgeInfo").Errorf("Error writing JSON over websocket: %v", err)
				break Top
			}
		case <-s.cancelBridge:
			break Top
		}
	}
	s.sendingBridgeInfo.UnSet()
}

func (s *StatPoller) stopBridgeInfo() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sendingBridgeInfo.IsSet() {
		s.cancelBridge <- struct{}{}
	}
}

// valueYoutubeInfo contains all information on youtube.Handler
type valueYoutubeInfo struct {
	NowPlaying        youtube.TrackInfo `json:"now_playing"`
	TrackDurationNs   int64             `json:"track_duration_ns"`
	ElapsedDurationNs int64             `json:"elapsed_duration_ns"`

	TrackIndex   int  `json:"track_index"`
	Paused       bool `json:"paused"`
	QueuedTracks []youtube.TrackInfo
}

func (s *StatPoller) sendYoutubeInfo(n int, interval time.Duration, ws *websocket.Conn) {
	if s.sendingYoutubeInfo.IsSet() {
		s.stopYoutubeInfo()
	}

	s.youtubeDispatchMu.Lock()
	defer s.youtubeDispatchMu.Unlock()

	s.sendingYoutubeInfo.Set()

	tick := time.NewTicker(interval)
	defer tick.Stop()

Top:
	for i := 0; i != n; i++ {
		select {
		case <-tick.C:
			nowPlaying, err := s.br.Controller().YouTube().NowPlaying()
			if err != nil {
				ws.WriteJSON(errorToJson(err))
				break Top
			}

			isPlaying, err := s.br.Controller().YouTube().IsPlaying()
			if err != nil {
				ws.WriteJSON(errorToJson(err))
				break Top
			}

			trackIndex, err := s.br.Controller().YouTube().TrackIndex()
			if err != nil {
				ws.WriteJSON(errorToJson(err))
				break Top
			}

			allTracks, err := s.br.Controller().YouTube().QueuedTracks()
			if err != nil {
				ws.WriteJSON(errorToJson(err))
				break Top
			}

			elapsed, err := s.br.Controller().YouTube().TimeElapsed()
			if err != nil {
				ws.WriteJSON(errorToJson(err))
				break Top
			}
			if err := ws.WriteJSON(&Response{
				Type:      RqtYoutubeInfo,
				Iteration: i,
				Value: &valueYoutubeInfo{
					NowPlaying:        nowPlaying,
					TrackDurationNs:   time.Duration(nowPlaying.Duration).Nanoseconds(),
					ElapsedDurationNs: elapsed.Nanoseconds(),
					TrackIndex:        trackIndex,
					Paused:            !isPlaying,
					QueuedTracks:      allTracks,
				},
			}); err != nil {
				log.Logger.WithField("category", "StatPoll BridgeInfo").Errorf("Error writing JSON over websocket: %v", err)
				break Top
			}
		case <-s.cancelYoutube:
			break Top
		}
	}

	s.sendingYoutubeInfo.UnSet()
}

func (s *StatPoller) stopYoutubeInfo() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sendingYoutubeInfo.IsSet() {
		s.cancelYoutube <- struct{}{}
	}
}

type valueAirPlayInfo struct {
	Server         *airplay2.Server   `json:"server"`
	Clients        []*airplay2.Client `json:"clients"`
	ArtworkURLPath string             `json:"artwork_url_path"`
}

func (s *StatPoller) sendAirPlayInfo(n int, interval time.Duration, ws *websocket.Conn) {
	if s.sendingAirPlayInfo.IsSet() {
		s.stopAirPlayInfo()
	}

	s.airplayDispatchMu.Lock()
	defer s.airplayDispatchMu.Unlock()

	s.sendingAirPlayInfo.Set()

	tick := time.NewTicker(interval)
	defer tick.Stop()

Top:
	for i := 0; i != n; i++ {
		select {
		case <-s.cancelAirPlay:
			break Top
		case <-tick.C:
			if err := ws.WriteJSON(&Response{
				Type:      RqtAirPlayInfo,
				Iteration: i,
				Value: &valueAirPlayInfo{
					Server:         s.br.Controller().AirPlay().Server(),
					Clients:        s.br.Controller().AirPlay().Clients(),
					ArtworkURLPath: "/api/bridge/artwork",
				},
			}); err != nil {
				log.Logger.WithField("category", "StatPoll AirPlayInfo").Errorf("Error writing JSON over websocket: %v", err)
				break Top
			}
		}
	}
	s.sendingAirPlayInfo.UnSet()
}

func (s *StatPoller) stopAirPlayInfo() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sendingAirPlayInfo.IsSet() {
		s.cancelAirPlay <- struct{}{}
		log.Logger.WithField("category", "StatPoll Dispatcher").Info("Stopping current AirPlay dispatcher...")
	}
}
