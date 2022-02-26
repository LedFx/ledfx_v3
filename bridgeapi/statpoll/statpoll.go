package statpoll

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/carterpeel/abool/v2"
	"github.com/gorilla/websocket"
	"io"
	"ledfx/audio/audiobridge"
	log "ledfx/logger"
	"sync"
	"time"
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

func (s *StatPoller) AddWebsocket(ws *websocket.Conn) error {
	if ws == nil {
		return errors.New("ws cannot be nil")
	}

	go s.socketLoop(ws)
	return nil
}

func (s *StatPoller) socketLoop(ws *websocket.Conn) {
	defer ws.Close()
	for {
		req := &Request{}
		if err := ws.ReadJSON(req); err != nil {
			switch {
			case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
				break
			case errors.As(err, &websocketClosedError):
				log.Logger.WithField("category", "StatPoll SocketLoop").Warnln("Websocket session closed")
				return
			case errors.As(err, &jsonSyntaxErr), errors.As(err, &jsonInvalidCharErr):
				log.Logger.WithField("category", "StatPoll SocketLoop").Warnln("Invalid request data received, skipping...")
				_ = ws.WriteJSON(errorToJson(err))
				continue
			default:
				log.Logger.WithField("category", "StatPoll SocketLoop").Errorf("Error reading JSON message from socket: %v", err)
				return
			}
		}
		if err := s.respondTo(ws, req); err != nil {
			log.Logger.WithField("category", "StatPoll SocketLoop").Errorf("Error initializing socket response: %v", err)
			return
		}
	}
}

func (s *StatPoller) respondTo(ws *websocket.Conn, r *Request) error {
Check:
	switch {
	case r == nil:
		fallthrough
	case ws == nil:
		return errors.New("(ws *websocket.Conn) and (r *Request) must be non-nil")
	case r.Type == "": // If Request.Type is unspecified, default to ReqBridgeInfo
		r.Type = ReqBridgeInfo
		goto Check
	case r.Iterations == 0:
		r.Iterations = 1
		goto Check
	case r.IntervalMs <= 0:
		r.IntervalMs = 1
	}

	switch r.Type {
	case ReqBridgeInfo:
		go s.sendBridgeInfo(r, time.Duration(r.IntervalMs)*time.Millisecond, ws)
	case ReqStopBridgeInfo:
		go s.stopBridgeInfo()
	case ReqYoutubeInfo:
		go s.sendYoutubeInfo(r, time.Duration(r.IntervalMs)*time.Millisecond, ws)
	case ReqStopYoutubeInfo:
		go s.stopYoutubeInfo()
	case ReqAirPlayInfo:
		go s.sendAirPlayInfo(r, time.Duration(r.IntervalMs)*time.Millisecond, ws)
	case ReqStopAirPlayInfo:
		go s.stopAirPlayInfo()
	default:
		return fmt.Errorf("unknown request type '%s'", r.Type)
	}
	return nil
}

type Request struct {
	Type       ReqType    `json:"type"`
	Params     []ReqParam `json:"params"`
	Iterations int        `json:"iterations"`
	IntervalMs int64      `json:"interval_ms"`
}

var (
	jsonSyntaxErr        *json.SyntaxError
	jsonInvalidCharErr   *json.InvalidUnmarshalError
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
