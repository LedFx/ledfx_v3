package utils

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

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
			log.Println(err)
			return
		}
		// print out that message for clarity
		fmt.Println(string(p))
		var msg Msg
		json.Unmarshal([]byte(p), &msg)
		// fmt.Printf("Type: %s, Message: %s", msg.Type, msg.Message)

		if msg.Message == "frontend connected" {
			// EXAMPLE DUMMY
			uptimeTicker := time.NewTicker(5 * time.Second)
			uptimeTickerb := time.NewTicker(5 * time.Second)
			dummyTypes := make([]string, 0)
			dummyTypes = append(dummyTypes,
				"success",
				"info",
				"warning",
				"error")
			dummyMsgs := make([]string, 0)
			dummyMsgs = append(dummyMsgs,
				"Sent from new LedFx-Go",
				"New core detected!",
				"BOOM",
				"Just like that")

			rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
			for {
				select {
				case <-uptimeTicker.C:
					if err := conn.WriteMessage(messageType, []byte(`{"type":"`+dummyTypes[rand.Intn(len(dummyTypes))]+`","message":"`+dummyMsgs[rand.Intn(len(dummyMsgs))]+`" }`)); err != nil {
						log.Println(err)
						return
					}
				case <-uptimeTickerb.C:

				}
			}
		}
	}
}

// define our WebSocket endpoint
func ServeWs(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(r.Host)

	// upgrade this connection to a WebSocket
	// connection
	ws, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	Reader(ws)
}
