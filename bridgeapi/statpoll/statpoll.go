package statpoll

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
	"io"
	"ledfx/audio/audiobridge"
	log "ledfx/logger"
	"time"
)

type StatPoller struct {
	br *audiobridge.Bridge

	sendingBridgeInfo  *atomic.Bool
	stopSendBridgeInfo *atomic.Bool

	sendingYoutubeInfo  *atomic.Bool
	stopSendYoutubeInfo *atomic.Bool
}

func New(br *audiobridge.Bridge) (s *StatPoller) {
	return &StatPoller{
		br:                  br,
		sendingBridgeInfo:   atomic.NewBool(false),
		stopSendBridgeInfo:  atomic.NewBool(false),
		sendingYoutubeInfo:  atomic.NewBool(false),
		stopSendYoutubeInfo: atomic.NewBool(false),
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
	case r.Type == "": // If Request.Type is unspecified, default to RqtBridgeInfo
		r.Type = RqtBridgeInfo
		goto Check
	case r.Iterations == 0:
		r.Iterations = 1
		goto Check
	case r.Iterations != 1 && r.IntervalMs <= 0:
		r.IntervalMs = 250
	}

	switch r.Type {
	case RqtBridgeInfo:
		go s.sendBridgeInfo(r.Iterations, time.Duration(r.IntervalMs)*time.Millisecond, ws)
	case RqtStopBridgeInfo:
		go s.stopBridgeInfo()
	case RqtYoutubeInfo:
		go s.sendYoutubeInfo(r.Iterations, time.Duration(r.IntervalMs)*time.Millisecond, ws)
	case RqtStopYoutubeInfo:
		go s.stopYoutubeInfo()
	default:
		return fmt.Errorf("unknown request type '%s'", r.Type)
	}
	return nil
}

type Request struct {
	Type       ReqType `json:"type"`
	Iterations int     `json:"iterations"`
	IntervalMs int64   `json:"interval_ms"`
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
