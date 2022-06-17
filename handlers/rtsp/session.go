package rtsp

import (
	"bytes"
	"fmt"
	log "ledfx/logger"
	"net"
	"strconv"
	"strings"

	"ledfx/handlers/sdp"
)

const (
	readBuffer = 2048
)

// Decrypter decrypts a received packet
type Decrypter interface {
	Decode([]byte) ([]byte, error)
}

// PortSet wraps the ports needed for an RTSP stream
type PortSet struct {
	Address string
	Control int
	Timing  int
	Data    int
}

// Session a streaming session
type Session struct {
	Description *sdp.SessionDescription
	decrypter   Decrypter
	RemotePorts PortSet
	LocalPorts  PortSet
	dataConn    net.Conn
	DataChan    chan []byte
	stopChan    chan struct{}
	buf         *bytes.Buffer

	sendBuf    []byte
	packetChan chan []byte
}

// NewSession instantiates a new Session
func NewSession(description *sdp.SessionDescription, decrypter Decrypter) *Session {
	return &Session{Description: description, decrypter: decrypter, DataChan: make(chan []byte, 1000), buf: bytes.NewBuffer(make([]byte, readBuffer)), sendBuf: make([]byte, 0), packetChan: make(chan []byte)}
}

// InitReceive initializes the session to for receiving
func (s *Session) InitReceive() error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", 0))
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	// keep track of the actual connection, so we can close it later
	s.dataConn = conn
	localAddr := strings.Split(conn.LocalAddr().String(), ":")
	s.LocalPorts.Data, _ = strconv.Atoi(localAddr[len(localAddr)-1])
	return nil
}

// Close closes a session
func (s *Session) Close(closeDone chan struct{}) {
	log.Logger.WithField("context", "RTSP Session").Infoln("Closing session...")
	s.stopChan = closeDone
	if s.dataConn != nil {
		s.dataConn.Close()
	} else {
		log.Logger.WithField("context", "RTSP Session").Infoln("Currently no data connection...")
		s.stopChan <- struct{}{}
	}
}

// StartReceiving starts a session for listening for data
func (s *Session) StartReceiving() error {
	// start listening for audio data
	log.Logger.WithField("context", "RTSP Session").Infoln("Session started. Listening for audio packets...")
	go func(conn *net.UDPConn) {
		for {
			n, _, err := conn.ReadFromUDP(s.buf.Bytes())
			if err != nil {
				log.Logger.WithField("context", "RTSP Session").Warnf("Error reading data from socket: %v", err)
				close(s.DataChan)
				conn = nil
				break
			}
			packet := s.buf.Bytes()[:n]
			// send the data to the decoder
			if s.decrypter != nil {
				if packet, err = s.decrypter.Decode(packet); err != nil {
					log.Logger.WithField("context", "RTSP Session").Errorf("Error decrypting packet: %v", err)
					return
				}
			}

			s.sendBuf = make([]byte, len(packet))
			copy(s.sendBuf, packet)
			s.DataChan <- s.sendBuf
		}
		log.Logger.WithField("context", "RTSP Session").Infoln("Signalling Session is closed")
		if s.stopChan != nil {
			s.stopChan <- struct{}{}
		}
	}(s.dataConn.(*net.UDPConn))
	return nil
}

// StartSending starts a session for sending data
func (s *Session) StartSending() (err error) {
	// keep track of the actual connection, so we can close it later.
	if s.dataConn, err = net.Dial("udp", fmt.Sprintf("%s:%d", s.RemotePorts.Address, s.RemotePorts.Data)); err != nil {
		return fmt.Errorf("error dialing '%s:%d': %w", s.RemotePorts.Address, s.RemotePorts.Data, err)
	}
	log.Logger.WithField("context", "RTSP Session").Infoln("Sending session started successfully")
	return nil
}

func (s *Session) DataConn() net.Conn {
	return s.dataConn
}
