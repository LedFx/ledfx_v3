package raop

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	log "ledfx/logger"
	"net"
	"strconv"
	"strings"
	"sync"

	"ledfx/handlers/player"
	"ledfx/handlers/rtsp"
	"ledfx/handlers/sdp"

	"github.com/grandcat/zeroconf"
)

// sets up the properties needed to make us discoverable as an airtunes service
// https://github.com/fgp/AirReceiver/blob/STABLE_1_X/src/main/java/org/phlo/AirReceiver/AirReceiver.java#L88
// https://nto.github.io/AirPlay.html#audio
const (
	airTunesServiceType = "_raop._tcp"
	domain              = "local."
	localTimingPort     = 6002
	localControlPort    = 6001
	headerTextParams    = "text/parameters"
)

var airtunesServiceProperties = []string{"txtvers=1",
	"tp=UDP",
	"ch=2",
	"ss=16",
	"sr=44100",
	"pw=false",
	"sm=false",
	"sv=false",
	"ek=1",
	"et=0,1",
	"cn=0,1",
	"md=0,1,2",
	"vn=3"}

// AirplayServer server for handling the RTSP protocol
type AirplayServer struct {
	port          int
	name          string
	rtspServer    *rtsp.Server
	zerconfServer *zeroconf.Server
	sessions      *sessionMap
	player        player.Player
	doneCh        chan struct{}
}

// Parameter types
type (
	ParamVolume    float64
	ParamTrackInfo struct {
		Album  string
		Artist string
		Title  string
	}
	ParamAlbumArt []byte
	ParamMuted    bool
)

type airplaySession struct {
	session *rtsp.Session
	client  *DacpClient
}

func newAirplaySession(session *rtsp.Session, dacpClient *DacpClient) *airplaySession {
	return &airplaySession{session, dacpClient}
}

type sessionMap struct {
	sync.RWMutex
	sessions map[string]*airplaySession
}

func newSessionMap() *sessionMap {
	return &sessionMap{sessions: make(map[string]*airplaySession)}
}

func (sm *sessionMap) addSession(address string, session *airplaySession) {
	sm.Lock()
	defer sm.Unlock()
	sm.sessions[address] = session
}

func (sm *sessionMap) removeSession(address string) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.sessions, address)
}

func (sm *sessionMap) getSession(address string) *airplaySession {
	sm.RLock()
	defer sm.RUnlock()
	s, ok := sm.sessions[address]
	if !ok {
		return nil
	}
	return s
}

func (sm *sessionMap) getSessions() []*airplaySession {
	sessions := make([]*airplaySession, 0, len(sm.sessions))

	for _, value := range sm.sessions {
		sessions = append(sessions, value)
	}
	return sessions
}

// NewAirplayServer instantiates a new airplayer server
func NewAirplayServer(port int, name string, player player.Player) *AirplayServer {
	as := AirplayServer{port: port, name: name, player: player, sessions: newSessionMap()}
	return &as
}

// Start starts the airplay server, broadcasting on bonjour, ready to accept requests
func (a *AirplayServer) Start(advertise bool) (err error) {
	if advertise {
		if err = a.initAdvertise(); err != nil {
			return err
		}
	}

	rtspServer := rtsp.NewServer(a.port)

	a.rtspServer = rtspServer

	rtspServer.AddHandler(rtsp.Options, handleOptions)
	rtspServer.AddHandler(rtsp.Announce, a.handleAnnounce)
	rtspServer.AddHandler(rtsp.Setup, a.handleSetup)
	rtspServer.AddHandler(rtsp.Record, a.handleRecord)
	rtspServer.AddHandler(rtsp.Set_Parameter, a.handleSetParameter)
	rtspServer.AddHandler(rtsp.Flush, handleFlush)
	rtspServer.AddHandler(rtsp.Teardown, a.handleTeardown)
	a.doneCh = make(chan struct{})
	rtspServer.Start(a.doneCh)
	return nil
}

func (a *AirplayServer) Wait() {
	<-a.doneCh
}

// ToggleAdvertise will toggle whether to advertise as an airplay service
func (a *AirplayServer) ToggleAdvertise(shouldAdvertise bool) (err error) {
	if !shouldAdvertise {
		if a.zerconfServer == nil {
			log.Logger.WithField("context", "RAOP Airplay").Println("Currently not advertising, ignoring turn off advertise request")
			return
		}
		// if we have a zerconfServer reference it means we are already advertising, so
		// stop it
		log.Logger.WithField("context", "RAOP Airplay").Printf("Shutting down broadcasting of %s", a.name)
		a.zerconfServer.Shutdown()
		a.zerconfServer = nil
	} else {
		if a.zerconfServer != nil {
			log.Logger.WithField("context", "RAOP Airplay").Println("Currently advertising, ignoring turn on advertise request")
			return nil
		}
		if err = a.initAdvertise(); err != nil {
			return err
		}
	}
	return nil
}

// ChangeName will change the name of the broadcast service
func (a *AirplayServer) ChangeName(newName string) (err error) {
	if strings.TrimSpace(newName) == "" {
		return errors.New("new name must be non-empty")
	}
	a.name = strings.TrimSpace(newName)
	// if we are advertising, stop the zeroconf server and start it so it
	// reflects the name change
	if a.zerconfServer != nil {
		a.zerconfServer.Shutdown()
		a.zerconfServer = nil
		if err = a.initAdvertise(); err != nil {
			return err
		}
	}
	return nil
}

func (a *AirplayServer) initAdvertise() (err error) {
	// as per the protocol, the mac address makes up part of the service name
	serviceName := fmt.Sprintf("%s@%s", strings.ReplaceAll(getMacAddr().String(), ":", ""), a.name)

	if a.zerconfServer, err = zeroconf.Register(serviceName, airTunesServiceType, domain, a.port, airtunesServiceProperties, nil); err != nil {
		return fmt.Errorf("failed to start ZeroConf server: %w", err)
	}

	log.Logger.WithField("context", "RAOP Advert").Println("Published service:")
	log.Logger.WithField("context", "RAOP Advert").Println("  - Name:", serviceName)
	log.Logger.WithField("context", "RAOP Advert").Println("  - Type:", airTunesServiceType)
	log.Logger.WithField("context", "RAOP Advert").Println("  - Domain:", domain)
	log.Logger.WithField("context", "RAOP Advert").Println("  - Port:", a.port)
	return nil
}

func handleOptions(req *rtsp.Request, resp *rtsp.Response, localAddress string, remoteAddress string) {
	resp.Status = rtsp.Ok
	resp.Headers["Public"] = strings.Join(rtsp.GetMethods(), " ")
	appleChallenge, exists := req.Headers["Apple-Challenge"]
	if !exists {
		return
	}
	log.Logger.WithField("context", "RAOP Handler: Options").Printf("Apple Challenge detected: %s\n", appleChallenge)
	challengResponse, err := generateChallengeResponse(appleChallenge, getMacAddr(), localAddress)
	if err != nil {
		log.Logger.WithField("context", "RAOP Handler: Options").Println("Error generating challenge response: ", err.Error())
	}
	resp.Headers["Apple-Response"] = challengResponse
}

func (a *AirplayServer) handleAnnounce(req *rtsp.Request, resp *rtsp.Response, localAddress string, remoteAddress string) {
	if req.Headers["Content-Type"] == "application/sdp" {
		description, err := sdp.Parse(bytes.NewReader(req.Body))
		if err != nil {
			log.Logger.WithField("context", "RAOP Handler: Announce").Println("error parsing SDP payload: ", err)
			resp.Status = rtsp.BadRequest
			return
		}
		// right now, we only maintain one audio session, so close any existing one
		a.closeAllSessions()
		var decoder rtsp.Decrypter

		if key, ok := description.Attributes["rsaaeskey"]; ok {
			aesKey, err := aeskeyFromRsa(key)
			if err != nil {
				log.Logger.WithField("context", "RAOP Handler: Announce").Println("Error retrieving aes key", err)
				resp.Status = rtsp.InternalServerError
				return
			}
			// from: https://github.com/joelgibson/go-airplay/blob/19e70c97e3903365f0a7f5a3f3c33751f4e8fb94/airplay/rtsp.go#L149
			aesIv64 := description.Attributes["aesiv"]
			aesIv64 = base64pad(aesIv64)
			aesIv, err := base64.StdEncoding.DecodeString(aesIv64)
			if err != nil {
				log.Logger.WithField("context", "RAOP Handler: Announce").Println("Error retrieving aes IV", err)
				resp.Status = rtsp.InternalServerError
				return
			}
			if decoder = NewAesDecrypter(aesKey, aesIv); err != nil {
				log.Logger.WithField("context", "RAOP Handler: Announce").Printf("Error creating new AES decryption object: %v\n", err)
				resp.Status = rtsp.InternalServerError
				return
			}
		}
		// create the dacp client for player control and then attach to the stream
		dacpID := req.Headers["DACP-ID"]
		activeRemote := req.Headers["Active-Remote"]
		dacpClient := DiscoverDacpClient(dacpID, activeRemote)
		s := rtsp.NewSession(description, decoder)
		if err = s.InitReceive(); err != nil {
			log.Logger.WithField("context", "RAOP Handler: Announce").Println("error initializing data receiving", err)
			resp.Status = rtsp.InternalServerError
			return
		}
		session := newAirplaySession(s, dacpClient)
		a.sessions.addSession(remoteAddress, session)
	}
	resp.Status = rtsp.Ok
}

func (a *AirplayServer) handleSetup(req *rtsp.Request, resp *rtsp.Response, localAddress string, remoteAddress string) {
	transport, hasTransport := req.Headers["Transport"]
	as := a.sessions.getSession(remoteAddress)
	if hasTransport {
		transportParts := strings.Split(transport, ";")
		var controlPort int
		var timingPort int
		for _, part := range transportParts {
			if strings.Contains(part, "control_port") {
				controlPort, _ = strconv.Atoi(strings.Split(part, "=")[1])
			}
			if strings.Contains(part, "timing_port") {
				timingPort, _ = strconv.Atoi(strings.Split(part, "=")[1])
			}
		}
		as.session.RemotePorts.Address = remoteAddress
		as.session.RemotePorts.Control = controlPort
		as.session.RemotePorts.Timing = timingPort
	}

	// hardcode our listening ports for now
	as.session.LocalPorts.Control = localControlPort
	as.session.LocalPorts.Timing = localTimingPort

	resp.Headers["Transport"] = fmt.Sprintf("RTP/AVP/UDP;unicast;mode=record;server_port=%d;control_port=%d;timing_port=%d", as.session.LocalPorts.Data, localControlPort, localTimingPort)
	resp.Headers["Session"] = "1"
	resp.Headers["Audio-Jack-Status"] = "connected"

	resp.Status = rtsp.Ok
}

func (a *AirplayServer) handleRecord(req *rtsp.Request, resp *rtsp.Response, localAddress string, remoteAddress string) {
	as := a.sessions.getSession(remoteAddress)
	err := as.session.StartReceiving()
	if err != nil {
		log.Logger.WithField("context", "RAOP Handler: Record").Println("could not start streaming session: ", err)
		resp.Status = rtsp.InternalServerError
		return
	}
	a.player.Play(as.session)
	resp.Headers["Audio-Latency"] = "2205"
	resp.Status = rtsp.Ok
}

func (a *AirplayServer) handleSetParameter(req *rtsp.Request, resp *rtsp.Response, localAddress string, remoteAddress string) {
	if req.Headers["Content-Type"] == "application/x-dmap-tagged" {
		daapData := parseDaap(req.Body)
		album := ""
		artist := ""
		title := ""
		if val, ok := daapData["daap.songalbum"]; ok {
			album = val.(string)
		}
		if val, ok := daapData["daap.songartist"]; ok {
			artist = val.(string)
		}
		if val, ok := daapData["dmap.itemname"]; ok {
			title = val.(string)
		}
		a.player.SetTrack(album, artist, title)
	} else if req.Headers["Content-Type"] == "image/jpeg" {
		a.player.SetAlbumArt(req.Body)
	} else if req.Headers["Content-Type"] == headerTextParams {
		body := string(req.Body)
		if strings.Contains(body, "volume") {
			volStr := strings.TrimSpace(strings.Split(body, "volume:")[1])
			vol, err := strconv.ParseFloat(volStr, 32)
			if err != nil {
				log.Logger.WithField("context", "RAOP Handler: SetParameter").Printf("Error converting volume to float: %v", err)
				resp.Status = rtsp.BadRequest
				return
			}
			if val, ok := req.Headers["X-BCG-Muted"]; ok {
				if val == "muted" {
					a.player.SetMute(true)
					// muting is enough, we don't need to bother
					// going on to set the actual volume
					resp.Status = rtsp.Ok
					return
				}
				a.player.SetMute(false)
			}
			a.player.SetVolume(normalizeVolume(vol))
		}
	}
	resp.Status = rtsp.Ok
}

func handleFlush(req *rtsp.Request, resp *rtsp.Response, localAddress string, remoteAddress string) {
	resp.Status = rtsp.Ok
}

func (a *AirplayServer) handleTeardown(req *rtsp.Request, resp *rtsp.Response, localAddress string, remoteAddress string) {
	a.closeSession(remoteAddress)
	resp.Status = rtsp.Ok
}

// Stop stops the airplay server
func (a *AirplayServer) Stop() {
	log.Logger.WithField("context", "AirPlay").Warnln("Stopping AirPlay server")
	a.closeAllSessions()
	a.rtspServer.Stop()
	if a.zerconfServer != nil {
		a.zerconfServer.Shutdown()
	}
}

func (a *AirplayServer) closeSession(remoteAddress string) {
	doneChan := make(chan struct{})
	as := a.sessions.getSession(remoteAddress)
	if as != nil {
		// stops the client from sending data
		if as.client != nil {
			as.client.Stop()
		}
		// closes the actual listening socket
		as.session.Close(doneChan)
		<-doneChan
		log.Logger.WithField("context", "RAOP Airplay").Println("Session closed")
		close(doneChan)
		a.sessions.removeSession(remoteAddress)
	}
}

func (a *AirplayServer) closeAllSessions() {
	for _, as := range a.sessions.getSessions() {
		a.closeSession(as.session.RemotePorts.Address)
	}
}

// getMacAddr gets the MAC hardware
// address of the host machine: https://gist.github.com/rucuriousyet/ab2ab3dc1a339de612e162512be39283
func getMacAddr() net.HardwareAddr {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && !bytes.Equal(i.HardwareAddr, nil) {
				// Don't use random as we have a real address
				return i.HardwareAddr
			}
		}
	}
	return nil
}

// normalizeVolume maps airplay volume values to a range between 0 and 1
func normalizeVolume(volume float64) float64 {
	// according to: https://nto.github.io/AirPlay.html#audio
	// -144 is mute
	// the remaining values will between -30 and 0,
	// so map that to a range between 0 and 1
	// simple range mapping formula: https://gamedev.stackexchange.com/questions/33441/how-to-convert-a-number-from-one-min-max-set-to-another-min-max-set
	// then simplified down and adjusted to make sure the number was positive

	switch volume {
	case -144:
		return 0
	case 0:
		return 1
	default:
		return (volume + 30) / 30
	}
}
