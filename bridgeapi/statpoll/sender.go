package statpoll

import (
	"errors"
	log "ledfx/logger"
	"ledfx/tickpool"
	"time"

	"github.com/gorilla/websocket"
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

		var err error

		for _, value := range statReq.Params {
			switch value {
			case ParamInputType:
				resp.Values[ParamInputType] = s.br.Controller().InputType()
			case ParamOutputs:
				resp.Values[ParamOutputs] = s.br.Controller().Outputs()
			case YtParamNowPlaying:
				resp.Values[YtParamNowPlaying], err = s.br.Controller().YouTube().NowPlaying()
			case YtParamTrackDuration:
				if nowPlaying, err := s.br.Controller().YouTube().NowPlaying(); err != nil {
					ws.WriteJSON(errorToJson(err))
					return
				} else {
					resp.Values[YtParamTrackDuration] = nowPlaying.Duration
				}
			case YtParamElapsedTime:
				resp.Values[YtParamElapsedTime], err = s.br.Controller().YouTube().TimeElapsed()
			case YtParamPaused:
				resp.Values[YtParamPaused], err = s.br.Controller().YouTube().IsPaused()
			case YtParamTrackIndex:
				resp.Values[YtParamTrackIndex], err = s.br.Controller().YouTube().TrackIndex()
			case YtParamQueuedTracks:
				resp.Values[YtParamQueuedTracks], err = s.br.Controller().YouTube().QueuedTracks()
			default:
				resp.Values[value] = "ERR_UNKNOWN_PARAMETER"
			}

			if err != nil {
				ws.WriteJSON(errorToJson(err))
				return
			}
		}
		select {
		case <-tick.C:
			if err := ws.WriteJSON(&resp); err != nil {
				if errors.As(err, &websocketClosedError) {
					log.Logger.WithField("context", "StatPoll SocketLoop").Warnln("Websocket session closed")
				} else {
					log.Logger.WithField("context", "StatPoll SocketLoop").Errorf("Error writing JSON to websocket: %v", err)
				}
				return
			}
		}
	}
}
