package utils

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// We'll need to define an Upgrader
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

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
type Msg struct {
	Type    string
	Message string
}

func Reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err, messageType)
			return
		}
		// print out that message for clarity
		fmt.Println(string(p))
		var msg Msg
		json.Unmarshal([]byte(p), &msg)

		// fmt.Printf("Type: %s, Message: %s", msg.Type, msg.Message)
		if msg.Message == "frontend connected" {
			SendWs(Ws, "info", "New Core detected!")
		}
	}
}

var Ws *websocket.Conn

// define our WebSocket endpoint
func ServeWs(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	var err error
	Ws, err = Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	Reader(Ws)
}

func SendWs(conn *websocket.Conn, msgType string, msg string) {
	if err := conn.WriteMessage(1, []byte(`{"type":"`+msgType+`","message":"`+msg+`" }`)); err != nil {
		log.Println(err)
		return
	}
}
