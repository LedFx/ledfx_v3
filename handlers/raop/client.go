package raop

import (
	"bytes"
	"fmt"
	log "ledfx/logger"
	"strconv"
	"strings"
	"time"

	"ledfx/handlers/rtsp"
	"ledfx/handlers/sdp"
)

// stateMachine that will handle the handshaking to set up the session with the
// client(s) we will be forwarding packets to
type stateFn func(client *rtsp.Client, session *rtsp.Session, codec CodecType) (stateFn, error)

type stateMachine struct {
	currentState stateFn
	codec        CodecType
}

func newStateMachine(codec CodecType) *stateMachine {
	return &stateMachine{currentState: initial, codec: codec}
}

func (sm *stateMachine) transition(client *rtsp.Client, session *rtsp.Session) (bool, error) {
	state, err := sm.currentState(client, session, sm.codec)
	if err != nil {
		return true, err
	}
	sm.currentState = state
	return sm.currentState != nil, err
}

type CodecType uint8

const (
	CodecTypeALAC CodecType = iota
	CodecTypePCM
)

// EstablishSession establishes a session that is ready to have data streamed through it
func EstablishSession(ip string, port int, codec CodecType) (*rtsp.Session, error) {
	client, err := rtsp.NewClient(ip, port)
	if err != nil {
		return nil, err
	}
	sessionDescription := sdp.NewSessionDescription()
	session := rtsp.NewSession(sessionDescription, nil)
	session.RemotePorts.Address = client.RemoteAddress()

	sm := newStateMachine(codec)
	handshaking := true

	for handshaking {
		handshaking, err = sm.transition(client, session)
		if err != nil {
			log.Logger.WithField("context", "RAOP Client").Println("Error encountered during RTSP handshaking, ", err)
			return nil, err
		}
	}
	log.Logger.WithField("context", "RAOP Client").Println("done handshaking")
	return session, nil
}

// our state functions below, emulating the airplay protocol, leaving
// out things like the encrypting and apple-challenge

// initial state, sends an OPTIONS to make sure all is up
func initial(client *rtsp.Client, session *rtsp.Session, codec CodecType) (stateFn, error) {
	req := rtsp.NewRequest()
	req.Method = rtsp.Options
	req.RequestURI = "*"
	resp, err := client.Send(req)
	if err != nil {
		return nil, err
	}
	if resp.Status != rtsp.Ok {
		return nil, fmt.Errorf("non-ok status returned: %s", resp.Status.String())
	}
	return announce, nil
}

func announce(client *rtsp.Client, session *rtsp.Session, codec CodecType) (stateFn, error) {
	req := rtsp.NewRequest()
	req.Method = rtsp.Announce
	sessionID := strconv.FormatInt(time.Now().Unix(), 10)
	localAddress := client.LocalAddress()
	req.RequestURI = fmt.Sprintf("rtsp://%s/%s", localAddress, sessionID)
	req.Headers["Content-Type"] = "application/sdp"
	// build the SDP payload
	sessionDescription := session.Description
	origin := sdp.Origin{}
	origin.AddrType = "IP4"
	origin.NetType = "IN"
	origin.SessionID = sessionID
	origin.SessionVersion = "0"
	origin.UnicastAddress = localAddress
	origin.Username = "bcp"
	sessionDescription.Origin = origin
	c := sdp.ConnectData{}
	c.AddrType = "IP4"
	c.NetType = "IN"
	c.ConnectionAddress = localAddress
	sessionDescription.ConnectData = c
	timing := sdp.Timing{StartTime: 0, StopTime: 0}
	sessionDescription.Timing = timing
	m := make([]sdp.MediaDescription, 1)
	md := sdp.MediaDescription{}
	md.Media = "audio"
	md.Fmt = "96"
	md.Port = "0"
	md.Proto = "RTP/AVP"
	m[0] = md
	sessionDescription.MediaDescription = m
	a := make(map[string]string)
	switch codec {
	case CodecTypePCM:
		a["rtpmap"] = "96 PCM"
	case CodecTypeALAC:
		a["rtpmap"] = "96 AppleLossless"
	}
	sessionDescription.Attributes = a
	// attach to request
	var b bytes.Buffer
	_, err := sdp.Write(&b, sessionDescription)
	if err != nil {
		log.Logger.WithField("context", "RAOP Client").Println("Error converting SDP payload, ", err)
		return nil, err
	}
	req.Body = b.Bytes()
	resp, err := client.Send(req)
	if err != nil {
		return nil, err
	}
	if resp.Status != rtsp.Ok {
		return nil, fmt.Errorf("non-ok status returned: %s", resp.Status.String())
	}
	return setup, nil
}

func setup(client *rtsp.Client, session *rtsp.Session, codec CodecType) (stateFn, error) {
	req := rtsp.NewRequest()
	req.Method = rtsp.Setup
	localAddress := client.LocalAddress()
	req.RequestURI = fmt.Sprintf("rtsp://%s/%s", localAddress, session.Description.Origin.SessionID)
	// hardcoded for now
	req.Headers["Transport"] = "RTP/AVP/UDP;unicast;interleaved=0-1;mode=record;control_port=8888;timing_port=8889"
	resp, err := client.Send(req)
	if err != nil {
		return nil, err
	}
	if resp.Status != rtsp.Ok {
		return nil, fmt.Errorf("non-ok status returned: %s", resp.Status.String())
	}
	transport := resp.Headers["Transport"]
	transportParts := strings.Split(transport, ";")
	var controlPort int
	var timingPort int
	var serverPort int
	for _, part := range transportParts {
		if strings.Contains(part, "control_port") {
			controlPort, _ = strconv.Atoi(strings.Split(part, "=")[1])
		}
		if strings.Contains(part, "timing_port") {
			timingPort, _ = strconv.Atoi(strings.Split(part, "=")[1])
		}
		if strings.Contains(part, "server_port") {
			serverPort, _ = strconv.Atoi(strings.Split(part, "=")[1])
		}
	}
	session.RemotePorts.Address = client.RemoteAddress()
	session.RemotePorts.Control = controlPort
	session.RemotePorts.Timing = timingPort
	session.RemotePorts.Data = serverPort

	return record, nil
}

func record(client *rtsp.Client, session *rtsp.Session, codecType CodecType) (stateFn, error) {
	req := rtsp.NewRequest()
	req.Method = rtsp.Record
	localAddress := client.LocalAddress()
	req.RequestURI = fmt.Sprintf("rtsp://%s/%s", localAddress, session.Description.Origin.SessionID)

	resp, err := client.Send(req)
	if err != nil {
		return nil, err
	}
	if resp.Status != rtsp.Ok {
		return nil, fmt.Errorf("non-ok status returned: %s", resp.Status.String())
	}
	return nil, nil
}
