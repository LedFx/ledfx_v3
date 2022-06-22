package websocket

import (
	"encoding/json"
	"ledfx/event"
	"ledfx/logger"
	"net/http"

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

func Serve(mux *http.ServeMux) {
	mux.HandleFunc("/websocket", New)
}

func New(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	logger.Logger.WithField("context", "Websocket").Debug("Creating websocket connection")
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.WithField("context", "Websocket").Warn(err)
	}
	ws := webSocket{
		conn: conn,
	}
	// subscribe to the events we want
	unsubLog := event.Subscribe(event.Log, ws.handleEvent)
	unsubRender := event.Subscribe(event.EffectRender, ws.handleEvent)
	defer unsubLog()
	defer unsubRender()
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	ws.Read()
}

type webSocket struct {
	conn *websocket.Conn
}

func (w *webSocket) handleEvent(e *event.Event) {
	w.Send(e)
}

func (w *webSocket) Send(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		logger.Logger.WithField("context", "Websocket").Error(err)
		return
	}
	err = w.conn.WriteMessage(1, b)
	if err != nil {
		logger.Logger.WithField("context", "Websocket").Error(err)
		return
	}
}

// Read will listen indefinitely for new messages
func (w *webSocket) Read() {
	for {
		// read in a message
		messageType, p, err := w.conn.ReadMessage()
		if err != nil {
			logger.Logger.WithField("context", "Websocket").Warn(err, messageType)
			return
		}
		// print out that message for clarity
		logger.Logger.Debug(string(p))
		// var msg Msg
		// err = json.Unmarshal([]byte(p), &msg)
		// if err != nil {
		// 	logger.Logger.WithField("context", "Websocket").Warn(err)
		// }
	}
}

// // SendWs will send a message to our WebSocket client
// func SendWs(conn *websocket.Conn, msgType string, msg string) {
// 	if err := conn.WriteMessage(1, []byte(`{"type":"`+msgType+`","message":"`+msg+`" }`)); err != nil {
// 		logger.Logger.WithField("context", "Websocket").Warn(err)
// 		return
// 	}
// }
