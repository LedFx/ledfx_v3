package statpoll

import (
	"encoding/json"
	"errors"
	"github.com/carterpeel/abool/v2"
	"github.com/gorilla/websocket"
	"ledfx/audio/audiobridge"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type StatPoller struct {
	mu *sync.RWMutex

	br *audiobridge.Bridge

	bridgeDispatchMu  *sync.Mutex
	sendingBridgeInfo *abool.AtomicBool
	cancelBridge      chan struct{}

	youtubeDispatchMu  *sync.Mutex
	sendingYoutubeInfo *abool.AtomicBool
	cancelYoutube      chan struct{}

	airplayDispatchMu  *sync.Mutex
	sendingAirPlayInfo *abool.AtomicBool
	cancelAirPlay      chan struct{}
}

func New(br *audiobridge.Bridge) (s *StatPoller) {
	return &StatPoller{
		br:                 br,
		mu:                 &sync.RWMutex{},
		sendingBridgeInfo:  abool.New(),
		cancelBridge:       make(chan struct{}),
		bridgeDispatchMu:   &sync.Mutex{},
		sendingYoutubeInfo: abool.New(),
		cancelYoutube:      make(chan struct{}),
		youtubeDispatchMu:  &sync.Mutex{},
		sendingAirPlayInfo: abool.New(),
		cancelAirPlay:      make(chan struct{}),
		airplayDispatchMu:  &sync.Mutex{},
	}
}

func (s *StatPoller) AddWebsocket(ws *websocket.Conn, r *http.Request) error {
	if ws == nil {
		return errors.New("ws cannot be nil")
	}

	statReq := Request{}

	var err error

	if intervalStr := r.URL.Query().Get("interval_ms"); intervalStr != "" {
		if statReq.IntervalMs, err = strconv.ParseInt(intervalStr, 10, 64); err != nil || statReq.IntervalMs < 1000 {
			statReq.IntervalMs = 1000
		}
	} else {
		statReq.IntervalMs = 1000
	}

	if paramsStr := r.URL.Query().Get("params"); paramsStr != "" {
		params := strings.Split(paramsStr, ",")
		statReq.Params = make([]ReqParam, len(params))
		for i := range params {
			statReq.Params[i] = ReqParam(params[i])
		}
	} else {
		statReq.Params = []ReqParam{
			ParamInputType,
			ParamOutputs,
			YtParamNowPlaying,
			YtParamTrackDuration,
			YtParamElapsedTime,
			YtParamPaused,
			YtParamTrackIndex,
			YtParamQueuedTracks,
		}
	}

	go s.socketLoop(ws, &statReq)
	return nil
}

type Request struct {
	Params     []ReqParam `json:"params"`
	IntervalMs int64      `json:"interval_ms"`
}

var (
	websocketClosedError *websocket.CloseError
)

type JsonError struct {
	Error string `json:"error"`
}

func errorToJson(err error) *JsonError {
	unwrapped := errors.Unwrap(err)
	if unwrapped != nil {
		err = unwrapped
	}
	return &JsonError{
		Error: err.Error(),
	}
}

func errToJsonBytes(err error) []byte {
	b, _ := json.Marshal(map[string]string{"error": err.Error()})
	return b
}
