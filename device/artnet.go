package device

import (
	"ledfx/color"
	"ledfx/logger"
	"net"
	"strconv"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

type ArtNet struct {
	Config     ArtNetConfig
	connection net.Conn
	pb         *packetBuilder
}

type ArtNetConfig struct {
	NetworkerConfig
	Universe int `mapstructure:"universe" json:"universe" description:"Starting universe for ArtDMX data" default:"0" validate:"gte=0,lte=16"`
	// TODO ArtSync?
}

func (d *ArtNet) initialize(base *Device, config map[string]interface{}) (err error) {
	defaults.Set(&d.Config)
	err = mapstructure.Decode(&config, &d.Config)
	if err != nil {
		return err
	}
	err = mapstructure.Decode(&config, &d.Config.NetworkerConfig)
	if err != nil {
		return err
	}
	err = validate.Struct(&d.Config)
	if err != nil {
		return err
	}
	d.pb, err = NewPacketBuilder(base.Config.PixelCount, ArtDMX, byte(d.Config.Universe)) // repurpose timeout for universe
	return err
}

func (d *ArtNet) send(p color.Pixels) (err error) {
	d.pb.Build(p)
	for i := range d.pb.packets {
		_, err = d.connection.Write(d.pb.packets[i])
	}
	return err
}

func (d *ArtNet) connect() (err error) {
	service := d.Config.IP + ":" + strconv.Itoa(d.Config.Port)
	remoteAddr, err := net.ResolveUDPAddr("udp", service)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		return err
	}
	d.connection = conn
	logger.Logger.Debugf("Established connection to %s \n", service)
	logger.Logger.Debugf("Remote UDP address : %s \n", conn.RemoteAddr().String())
	logger.Logger.Debugf("Local UDP client address : %s \n", conn.LocalAddr().String())
	return nil
}

func (d *ArtNet) disconnect() error {
	return d.connection.Close()
}
