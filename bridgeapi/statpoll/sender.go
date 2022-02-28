package statpoll

import (
	"github.com/gorilla/websocket"
	"ledfx/audio/audiobridge"
	"ledfx/audio/audiobridge/youtube"
	"ledfx/integrations/airplay2"
	log "ledfx/logger"
	"ledfx/tickpool"
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

func (s *StatPoller) sendBridgeInfo(r *Request, interval time.Duration, ws *websocket.Conn) {
	if s.sendingBridgeInfo.IsSet() {
		s.stopBridgeInfo()
	}

	s.bridgeDispatchMu.Lock()
	defer s.bridgeDispatchMu.Unlock()

	s.sendingBridgeInfo.Set()

	tick := tickpool.Get(interval)
	defer tickpool.Put(tick)

Top:
	for i := 0; i != r.Iterations; i++ {
		select {
		case <-tick.C:
			if err := ws.WriteJSON(&Response{
				Type:      ReqBridgeInfo,
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
	NowPlaying        *youtube.TrackInfo `json:"now_playing,omitempty"`
	TrackDurationNs   int64              `json:"track_duration_ns,omitempty"`
	ElapsedDurationNs int64              `json:"elapsed_duration_ns,omitempty"`

	TrackIndex   int                 `json:"track_index,omitempty"`
	Paused       bool                `json:"paused,omitempty"`
	QueuedTracks []youtube.TrackInfo `json:"queued_tracks,omitempty"`
}

func (s *StatPoller) sendYoutubeInfo(r *Request, interval time.Duration, ws *websocket.Conn) {
	if s.sendingYoutubeInfo.IsSet() {
		s.stopYoutubeInfo()
	}

	s.youtubeDispatchMu.Lock()
	defer s.youtubeDispatchMu.Unlock()

	s.sendingYoutubeInfo.Set()

	tick := tickpool.Get(interval)
	defer tickpool.Put(tick)

Top:
	for i := 0; i != r.Iterations; i++ {
		select {
		case <-tick.C:
			if len(r.Params) <= 0 {
				r.Params = []ReqParam{YtParamNowPlaying, YtParamTrackDuration, YtParamElapsedTime, YtParamPaused, YtParamTrackIndex, YtParamQueuedTracks}
			}

			nowPlaying, err := s.br.Controller().YouTube().NowPlaying()
			if err != nil {
				ws.WriteJSON(errorToJson(err))
				break Top
			}
			value := new(valueYoutubeInfo)

			for i2 := range r.Params {
				switch r.Params[i2] {
				case YtParamNowPlaying:
					value.NowPlaying = &nowPlaying
				case YtParamTrackDuration:
					value.TrackDurationNs = time.Duration(nowPlaying.Duration).Nanoseconds()
				case YtParamElapsedTime:
					elapsed, err := s.br.Controller().YouTube().TimeElapsed()
					if err != nil {
						ws.WriteJSON(errorToJson(err))
						break Top
					}

					if elapsed == 0 {
						value.ElapsedDurationNs = -1
					} else {
						value.ElapsedDurationNs = elapsed.Nanoseconds()
					}
				case YtParamPaused:
					if value.Paused, err = s.br.Controller().YouTube().IsPaused(); err != nil {
						ws.WriteJSON(errorToJson(err))
						break Top
					}
				case YtParamTrackIndex:
					if value.TrackIndex, err = s.br.Controller().YouTube().TrackIndex(); err != nil {
						ws.WriteJSON(errorToJson(err))
						break Top
					}
				case YtParamQueuedTracks:
					if value.QueuedTracks, err = s.br.Controller().YouTube().QueuedTracks(); err != nil {
						ws.WriteJSON(errorToJson(err))
						break Top
					}
				}
			}
			if err = ws.WriteJSON(&Response{
				Type:      ReqYoutubeInfo,
				Iteration: i,
				Value:     value,
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

func (s *StatPoller) sendAirPlayInfo(r *Request, interval time.Duration, ws *websocket.Conn) {
	if s.sendingAirPlayInfo.IsSet() {
		s.stopAirPlayInfo()
	}

	s.airplayDispatchMu.Lock()
	defer s.airplayDispatchMu.Unlock()

	s.sendingAirPlayInfo.Set()

	tick := tickpool.Get(interval)
	defer tickpool.Put(tick)

Top:
	for i := 0; i != r.Iterations; i++ {
		select {
		case <-s.cancelAirPlay:
			break Top
		case <-tick.C:
			if err := ws.WriteJSON(&Response{
				Type:      ReqAirPlayInfo,
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
