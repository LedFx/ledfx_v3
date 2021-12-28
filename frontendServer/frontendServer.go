//go:generate goversioninfo -icon=assets/logo.ico
package frontendServer

import (
	"archive/zip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"github.com/gorilla/websocket"
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
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

func reader(conn *websocket.Conn) {
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
func serveWs(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(r.Host)

	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	reader(ws)
}

func unzip() {
	dst := "dest"
	archive, err := zip.OpenReader("new_frontend.zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)
		// fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return
		}
		if f.FileInfo().IsDir() {
			// fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}
	os.Rename("./dest/ledfx_frontend_v2", "./frontend")

	// cleanup
	if _, err := os.Stat("dest"); err == nil {
		os.RemoveAll("dest")
	}

	defer os.RemoveAll("new_frontend.zip")

}

func serveFrontend() {
	serveFrontend := http.FileServer(http.Dir("frontend"))
	fileMatcher := regexp.MustCompile(`\.[a-zA-Z]*$`)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !fileMatcher.MatchString(r.URL.Path) {
			http.ServeFile(w, r, "frontend/index.html")
		} else {
			serveFrontend.ServeHTTP(w, r)
		}
	})
}

func getFrontend() {
	log.Println("Getting latest Frontend")
	resp, err := http.Get("https://github.com/YeonV/LedFx-Frontend-v2/releases/latest/download/ledfx_frontend_v2.zip")
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}
	// Delete old files
	if _, err := os.Stat("frontend"); err == nil {
		os.RemoveAll("frontend")
	}
	defer os.RemoveAll("new_frontend.zip")

	// Create the file
	out, err := os.Create("new_frontend.zip")
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	// Extract frontend
	unzip()
	log.Println("Got latest Frontend")
	fmt.Println("===========================================")
}

func setupRoutes() {
	getFrontend()
	serveFrontend()
	// map our `/ws` endpoint to the `serveWs` function
	http.HandleFunc("/ws", serveWs)
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

//go:embed assets/logo.ico
var logo []byte

func onReady() {
	systray.SetIcon(logo)
	systray.SetTitle("LedFx-Go")
	systray.SetTooltip("LedFx-Go")
	mOpen := systray.AddMenuItem("Open", "Open LedFx in Browser")
	mGithub := systray.AddMenuItem("Github", "Open LedFx in Browser")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")
	for {
		select {
		case <-mOpen.ClickedCh:
			openbrowser("http://localhost:8080")
		case <-mGithub.ClickedCh:
			openbrowser("https://github.com/YeonV/ledfx-go")
		case <-mQuit.ClickedCh:
			systray.Quit()
		}
	}
}

func InitFrontend() {
	fmt.Println("===========================================")
	fmt.Println("          LedFx-Frontend by Blade")
	fmt.Println("    [CTRL]+Click: http://localhost:8080")
	fmt.Println("===========================================")
	setupRoutes()
	go func() {
		http.ListenAndServe(":8080", nil)
	}()
	systray.Run(onReady, nil)
}
