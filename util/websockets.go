package util

import (
	"encoding/json"
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

// Msg defines a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
type Msg struct {
	Type    string
	Message string
}

// Reader will listen indefinitely for new messages
func Reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			logger.Logger.Warn(err, messageType)
			return
		}
		// print out that message for clarity
		logger.Logger.Debug(string(p))
		var msg Msg
		err = json.Unmarshal([]byte(p), &msg)
		if err != nil {
			logger.Logger.Warn(err)
		}

		// fmt.Printf("Type: %s, Message: %s", msg.Type, msg.Message)
		if msg.Message == "frontend connected" {
			SendWs(Ws, "info", "New Core detected!")
		}
	}
}

// Ws is our global websocket connection
var Ws *websocket.Conn

// TODO: Handle more than one connection

func ServeWebsocket() {
	http.HandleFunc("/ws", ServeWs)
}

// ServeWs defines our WebSocket endpoint
func ServeWs(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	var err error
	Ws, err = Upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.Warn(err)
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	Reader(Ws)
}

// SendWs will send a message to our WebSocket client
func SendWs(conn *websocket.Conn, msgType string, msg string) {
	if err := conn.WriteMessage(1, []byte(`{"type":"`+msgType+`","message":"`+msg+`" }`)); err != nil {
		logger.Logger.Warn(err)
		return
	}
}
