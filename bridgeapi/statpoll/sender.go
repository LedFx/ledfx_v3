package statpoll

import (
	"errors"
	"github.com/gorilla/websocket"
	log "ledfx/logger"
	"ledfx/tickpool"
	"time"
)

type Response struct {
	Values map[ReqParam]interface{} `json:"values"`
}

func (s *StatPoller) socketLoop(ws *websocket.Conn, statReq *Request) {
	defer ws.Close()

	tick := tickpool.Get(time.Duration(statReq.IntervalMs) * time.Millisecond)
	defer tickpool.Put(tick)

	for {
		resp := Response{
			Values: make(map[ReqParam]interface{}),
		}

		if err := s.br.Controller().YouTube().CheckErr(); err != nil {
			ws.WriteJSON(errorToJson(err))
			return
		}

		for _, value := range statReq.Params {
			switch value {
			case YtParamNowPlaying:
				resp.Values[YtParamNowPlaying], _ = s.br.Controller().YouTube().NowPlaying()
			case YtParamTrackDuration:
				nowPlaying, _ := s.br.Controller().YouTube().NowPlaying()
				resp.Values[YtParamTrackDuration] = nowPlaying.Duration
			case YtParamElapsedTime:
				resp.Values[YtParamElapsedTime], _ = s.br.Controller().YouTube().TimeElapsed()
			case YtParamPaused:
				resp.Values[YtParamPaused], _ = s.br.Controller().YouTube().IsPaused()
			case YtParamTrackIndex:
				resp.Values[YtParamTrackIndex], _ = s.br.Controller().YouTube().TrackIndex()
			case YtParamQueuedTracks:
				resp.Values[YtParamQueuedTracks], _ = s.br.Controller().YouTube().QueuedTracks()
			default:
				resp.Values[value] = "ERR_UNKNOWN_PARAMETER"
			}
		}
		select {
		case <-tick.C:
			if err := ws.WriteJSON(&resp); err != nil {
				if errors.As(err, &websocketClosedError) {
					log.Logger.WithField("category", "StatPoll SocketLoop").Warnln("Websocket session closed")
				} else {
					log.Logger.WithField("category", "StatPoll SocketLoop").Errorf("Error writing JSON to websocket: %v", err)
				}
				return
			}
		}
	}
}
