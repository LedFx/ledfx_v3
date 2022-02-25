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
	if s.sendingBridgeInfo.Load() {
		s.stopBridgeInfo()
		time.Sleep(interval)
	}

	s.sendingBridgeInfo.Store(true)
	defer s.sendingBridgeInfo.Store(false)

	info := s.br.Info()
	for i := 0; i != n; i++ {

		now := time.Now()

		resp := &Response{
			Type:      RqtBridgeInfo,
			Iteration: i,
			Value: &valueBridgeInfo{
				InputType: info.InputType(),
				Outputs:   info.AllOutputs(),
			},
		}

		if s.stopSendBridgeInfo.Load() {
			s.stopSendBridgeInfo.Store(false)
			return
		}

		if err := ws.WriteJSON(resp); err != nil {
			log.Logger.WithField("category", "StatPoll BridgeInfo").Errorf("Error writing JSON over websocket: %v", err)
			return
		}
		time.Sleep(interval - time.Since(now))
	}
}

func (s *StatPoller) stopBridgeInfo() {
	if s.sendingBridgeInfo.Load() {
		s.stopSendBridgeInfo.Store(true)
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
	if s.sendingYoutubeInfo.Load() {
		s.stopYoutubeInfo()
		time.Sleep(interval)
	}

	s.sendingYoutubeInfo.Store(true)
	defer s.sendingYoutubeInfo.Store(false)

	yt := s.br.Controller().YouTube()
	for i := 0; i != n; i++ {
		now := time.Now()

		nowPlaying, err := yt.NowPlaying()
		if err != nil {
			ws.WriteJSON(errorToJson(err))
			return
		}

		isPlaying, err := yt.IsPlaying()
		if err != nil {
			ws.WriteJSON(errorToJson(err))
			return
		}

		trackIndex, err := yt.TrackIndex()
		if err != nil {
			ws.WriteJSON(errorToJson(err))
			return
		}

		allTracks, err := yt.QueuedTracks()
		if err != nil {
			ws.WriteJSON(errorToJson(err))
			return
		}

		elapsed, err := yt.TimeElapsed()
		if err != nil {
			ws.WriteJSON(errorToJson(err))
			return
		}

		resp := &Response{
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
		}

		if s.stopSendYoutubeInfo.Load() {
			s.stopSendYoutubeInfo.Store(false)
			return
		}
		if err := ws.WriteJSON(resp); err != nil {
			log.Logger.WithField("category", "StatPoll BridgeInfo").Errorf("Error writing JSON over websocket: %v", err)
			return
		}
		time.Sleep(interval - time.Since(now))
	}
}

func (s *StatPoller) stopYoutubeInfo() {
	if s.sendingYoutubeInfo.Load() {
		s.stopSendYoutubeInfo.Store(true)
	}
}

type valueAirPlayInfo struct {
	Server         *airplay2.Server   `json:"server"`
	Clients        []*airplay2.Client `json:"clients"`
	ArtworkURLPath string             `json:"artwork_url_path"`
}

func (s *StatPoller) sendAirPlayInfo(n int, interval time.Duration, ws *websocket.Conn) {
	if s.sendingAirPlayInfo.Load() {
		s.stopAirPlayInfo()
		time.Sleep(interval)
	}

	s.sendingAirPlayInfo.Store(true)
	defer s.sendingAirPlayInfo.Store(false)

	ap := s.br.Controller().AirPlay()
	for i := 0; i != n; i++ {
		now := time.Now()

		resp := &Response{
			Type:      RqtAirPlayInfo,
			Iteration: i,
			Value: &valueAirPlayInfo{
				Server:         ap.Server(),
				Clients:        ap.Clients(),
				ArtworkURLPath: "/api/bridge/artwork",
			},
		}

		if s.stopSendAirPlayInfo.Load() {
			s.stopSendAirPlayInfo.Store(false)
			return
		}
		if err := ws.WriteJSON(resp); err != nil {
			log.Logger.WithField("category", "StatPoll AirPlayInfo").Errorf("Error writing JSON over websocket: %v", err)
			return
		}
		time.Sleep(interval - time.Since(now))
	}
}

func (s *StatPoller) stopAirPlayInfo() {
	if s.sendingAirPlayInfo.Load() {
		s.stopSendAirPlayInfo.Store(true)
	}
}
