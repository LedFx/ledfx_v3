package websocket

import (
	"encoding/json"
	"ledfx/event"
	"ledfx/logger"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Upgrader will need to be defined
// this will require a Read and Write buffer size
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// We'll need to check the origin of our connection
	// this will allow us to make requests from our React
	// development server to here.
	// For now, we'll do no checking and just allow any connection
	CheckOrigin: func(r *http.Request) bool { return true },
}

type webSocket struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *webSocket) handleEvent(e *event.Event) {
	w.Send(e)
}

func (w *webSocket) Send(v any) {
	w.mu.Lock()
	defer w.mu.Unlock()
	b, err := json.Marshal(v)
	if err != nil {
		logger.Logger.WithField("context", "Websocket").Debug(err)
		return
	}
	err = w.conn.WriteMessage(1, b)
	if err != nil {
		logger.Logger.WithField("context", "Websocket").Debug(err)
		return
	}
}

// Read will listen indefinitely for new messages
func (w *webSocket) Read() {
	for {
		// read in a message
		messageType, p, err := w.conn.ReadMessage()
		if err != nil {
			logger.Logger.WithField("context", "Websocket").Debug(err, messageType)
			return
		}
		// print out that message for clarity
		logger.Logger.WithField("context", "Websocket").Debug(string(p))
		// TODO websockets API
	}
}

func Serve(mux *http.ServeMux) {
	mux.HandleFunc("/websocket", New)
}

func New(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket connection
	logger.Logger.WithField("context", "Websocket").Debugf("Creating connection with %s", r.RemoteAddr)
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.WithField("context", "Websocket").Error(err)
		return
	}
	ws := webSocket{
		conn: conn,
		mu:   sync.Mutex{},
	}
	logger.Logger.WithField("context", "Websocket").Debugf("Connection established with %s", r.RemoteAddr)
	// subscribe to the events we want
	// there are eight event types so we'll just ask for all of them
	var i event.EventType
	for i = 0; i <= 11; i++ {
		// sub and also defer calling the unsubscribe function
		defer event.Subscribe(i, ws.handleEvent)()
	}
	// listen indefinitely for new messages coming through on our WebSocket connection
	ws.Read()
	logger.Logger.WithField("context", "Websocket").Debugf("Closed connection with %s", r.RemoteAddr)
}
