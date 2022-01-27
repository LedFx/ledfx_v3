package airplay2

import (
	"fmt"
	"github.com/grantmd/go-airplay"
	"ledfx/handlers/raop"
	"ledfx/handlers/rtsp"
	log "ledfx/logger"
	"net"
	"strconv"
	"time"
)

type Client struct {
	dev       *airplay.AirplayDevice
	session   *rtsp.Session
	DataConn  net.Conn
	paramConn *rtsp.Client
}

type ClientDiscoveryParameters struct {
	// DeviceNameRegex ignores case and will connect to any device that contains this string.
	DeviceNameRegex string

	// DeviceIP takes higher priority than DeviceNameRegex if it is populated.
	DeviceIP string

	// Verbose, when true, enables the printing of all discovered devices to stderr.
	Verbose bool
}

func NewClient(searchParameters ClientDiscoveryParameters) (cl *Client, err error) {
	device, err := queryDevice(searchParameters)
	if err != nil {
		return nil, fmt.Errorf("error querying for device by name: %w", err)
	}

	log.Logger.WithField("category", "AirPlay Client").Infof("Establishing session with %s...", device.Name)
	session, err := raop.EstablishSession(device.IP.String(), int(device.Port))
	if err != nil {
		return nil, fmt.Errorf("error establishing RTSP session: %w", err)
	}

	if err := session.StartSending(); err != nil {
		return nil, fmt.Errorf("error during session.StartSending(): %w", err)
	}

	paramConn, err := rtsp.NewClient(device.IP.String(), int(device.Port))
	if err != nil {
		log.Logger.WithField("category", "AirPlay Client").Errorf("Error establishing RTSP session: %v\n", err)
		return
	}

	cl = &Client{
		dev:       device,
		session:   session,
		paramConn: paramConn,
		DataConn:  session.DataConn(),
	}

	return cl, nil
}

func (cl *Client) SetParam(par interface{}) {
	var err error
	req := rtsp.NewRequest()
	req.Method = rtsp.Set_Parameter
	req.RequestURI = fmt.Sprintf("rtsp://%s/%s", cl.paramConn.LocalAddress(), strconv.FormatInt(time.Now().Unix(), 10))

	switch val := par.(type) {
	case raop.ParamVolume:
		log.Logger.WithField("category", "AirPlay Client").Infof("Propagating volume '%f' to destination server\n", val)

		req.Headers["Content-Type"] = "text/parameters"
		req.Body = []byte(fmt.Sprintf("volume: %f", val))
	case raop.ParamMuted:
		log.Logger.WithField("category", "AirPlay Client").Infof("Propagating muted value '%v' to destination server\n", val)

		req.Headers["Content-Type"] = "text/parameters"
		req.Body = []byte("volume: -144")
	case raop.ParamTrackInfo:
		log.Logger.WithField("category", "AirPlay Client").Infof("Propagating track info to destination server\n")

		req.Headers["Content-Type"] = "application/x-dmap-tagged"
		if req.Body, err = raop.EncodeDaap(map[string]interface{}{"daap.songalbum": val.Album, "daap.itemname": val.Title, "daap.songartist": val.Artist}); err != nil {
			log.Logger.WithField("category", "AirPlay Client").Errorf("Error encoding track information as DAAP: %v\n", err)
			return
		}
	case raop.ParamAlbumArt:
		log.Logger.WithField("category", "AirPlay Client").Infof("Propagating album art to destination server\n")
		req.Headers["Content-Type"] = "image/jpeg"
		req.Body = val
	}
	resp, err := cl.paramConn.Send(req)
	if err != nil {
		log.Logger.WithField("category", "AirPlay Client").Errorf("Error sending volume request: %v\n", err)
		return
	}
	if resp.Status != rtsp.Ok {
		log.Logger.WithField("category", "AirPlay Client").Errorf("Unexpected response code from destination server: %s\n", resp.Status.String())
		log.Logger.WithField("category", "AirPlay Client").Errorf("Response: \n----\n[%s]\n----\n", resp.String())
	}
}

func (cl *Client) Write(p []byte) (n int, err error) {
	return cl.DataConn.Write(p)
}

func (cl *Client) Close() {
	if cl.session != nil {
		_ = cl.DataConn.Close()
	}
}
