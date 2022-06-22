package airplay2

import (
	"encoding/json"
	"fmt"
	"ledfx/handlers/raop"
	"ledfx/handlers/rtsp"
	log "ledfx/logger"
	"net"
	"time"

	"github.com/grantmd/go-airplay"
)

type Client struct {
	dev          *airplay.AirplayDevice
	session      *rtsp.Session
	DataConn     net.Conn
	paramConn    *rtsp.Client
	searchParams *ClientDiscoveryParameters
}

type ClientDiscoveryParameters struct {
	// DeviceNameRegex ignores case and will connect to any device that contains this string.
	DeviceNameRegex string

	// DeviceIP takes higher priority than DeviceNameRegex if it is populated.
	DeviceIP string
}

func NewClient(searchParameters ClientDiscoveryParameters) (cl *Client, err error) {
	device, err := queryDevice(searchParameters)
	if err != nil {
		return nil, fmt.Errorf("error querying for device by name: %w", err)
	}

	cl = &Client{
		dev:          device,
		searchParams: &searchParameters,
	}

	return cl, nil
}

func (cl *Client) ConfirmConnect() (err error) {
	log.Logger.WithField("context", "AirPlay Client").Infof("Establishing session with %s...", cl.dev.Name)
	if cl.session, err = raop.EstablishSession(cl.dev.IP.String(), int(cl.dev.Port), raop.CodecTypePCM); err != nil {
		return fmt.Errorf("error establishing RTSP session: %w", err)
	}

	if err := cl.session.StartSending(); err != nil {
		return fmt.Errorf("error during session.StartSending(): %w", err)
	}

	if cl.paramConn, err = rtsp.NewClient(cl.dev.IP.String(), int(cl.dev.Port)); err != nil {
		return fmt.Errorf("error establishing RTSP session: %w", err)
	}

	cl.DataConn = cl.session.DataConn()
	return nil
}

func (cl *Client) Name() string {
	return cl.dev.Name
}
func (cl *Client) Hostname() string {
	return cl.dev.Hostname
}
func (cl *Client) RemoteIP() net.IP {
	return cl.dev.IP
}
func (cl *Client) RemotePort() int {
	return int(cl.dev.Port)
}
func (cl *Client) Type() string {
	return cl.dev.Type
}
func (cl *Client) DeviceModel() string {
	return cl.dev.DeviceModel()
}
func (cl *Client) SampleRate() int {
	return cl.dev.AudioSampleRate()
}

func (cl *Client) WriterID() string {
	return cl.session.RemotePorts.Address
}

func (cl *Client) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Name        string `json:"name"`
		Hostname    string `json:"hostname"`
		RemoteIP    string `json:"remote_ip"`
		RemotePort  int    `json:"remote_port"`
		Type        string `json:"type"`
		DeviceModel string `json:"device_model"`
		SampleRate  int    `json:"sample_rate"`
		WriterID    string `json:"writer_id"`
	}{
		Name:        cl.Name(),
		Hostname:    cl.Hostname(),
		RemoteIP:    cl.RemoteIP().To4().String(),
		RemotePort:  cl.RemotePort(),
		Type:        cl.Type(),
		DeviceModel: cl.DeviceModel(),
		SampleRate:  cl.SampleRate(),
		WriterID:    cl.WriterID(),
	})
}

func (cl *Client) SetParam(par interface{}) {
	switch {
	case cl == nil:
		return
	case cl.paramConn == nil:
		return
	}

	var err error
	req := rtsp.NewRequest()
	req.Method = rtsp.Set_Parameter
	req.RequestURI = fmt.Sprintf("rtsp://%s/%d", cl.paramConn.LocalAddress(), time.Now().Unix())

	switch val := par.(type) {
	case raop.ParamVolume:
		log.Logger.WithField("context", "AirPlay Client").Infof("Propagating volume '%f' to destination server", val)

		req.Headers["Content-Type"] = "text/parameters"
		req.Body = []byte(fmt.Sprintf("volume: %f", val))
	case raop.ParamMuted:
		log.Logger.WithField("context", "AirPlay Client").Infof("Propagating muted value '%v' to destination server", val)

		req.Headers["Content-Type"] = "text/parameters"
		req.Body = []byte("volume: -144")
	case raop.ParamTrackInfo:
		log.Logger.WithField("context", "AirPlay Client").Infof("Propagating track info to destination server")

		req.Headers["Content-Type"] = "application/x-dmap-tagged"
		if req.Body, err = raop.EncodeDaap(map[string]interface{}{"daap.songalbum": val.Album, "daap.itemname": val.Title, "daap.songartist": val.Artist}); err != nil {
			log.Logger.WithField("context", "AirPlay Client").Errorf("Error encoding track information as DAAP: %v\n", err)
			return
		}
	case raop.ParamAlbumArt:
		log.Logger.WithField("context", "AirPlay Client").Infof("Propagating album art to destination server\n")
		req.Headers["Content-Type"] = "image/jpeg"
		req.Body = val
	}

	if cl.paramConn == nil {
		log.Logger.WithField("context", "AirPlay Client").Warnf("Rebuilding RTSP parameter session...")
		if cl.paramConn, err = rtsp.NewClient(cl.dev.IP.String(), int(cl.dev.Port)); err != nil {
			log.Logger.WithField("context", "AirPlay Client").Errorf("Error establishing RTSP parameter session: %v", err)
		}
	}

	resp, err := cl.paramConn.Send(req)
	if err != nil {
		log.Logger.WithField("context", "AirPlay Client").Errorf("Error sending volume request: %v\n", err)
		return
	}
	if resp.Status != rtsp.Ok {
		log.Logger.WithField("context", "AirPlay Client").Errorf("Unexpected response code from destination server: %s\n", resp.Status.String())
		log.Logger.WithField("context", "AirPlay Client").Errorf("Response: \n----\n[%s]\n----\n", resp.String())
	}
}

func (cl *Client) Close() {
	if cl.session != nil {
		if err := cl.session.DataConn().Close(); err != nil {
			log.Logger.WithField("context", "AirPlay Client").Errorf("Error closing client session conn: %v", err)
		}
	}
	if cl.paramConn != nil {
		if err := cl.paramConn.Close(); err != nil {
			log.Logger.WithField("context", "AirPlay Client").Errorf("Error closing client param conn: %v", err)
		}
		cl.paramConn = nil
	}
}

func (cl *Client) Write(p []byte) (int, error) {
	return cl.DataConn.Write(p)
}
