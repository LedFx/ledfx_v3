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
	config     ArtNetConfig
	connection net.Conn
	pb         *packetBuilder
}

type ArtNetConfig struct {
	IP       string `mapstructure:"ip" json:"ip" description:"Device IP address on the LAN" validate:"required,ip"`
	Port     int    `mapstructure:"port" json:"port" description:"Port number the device is listening on" default:"21324" validate:"gte=0,lte=65535"`
	Universe int    `mapstructure:"universe" json:"universe" description:"Starting universe for ArtDMX data" default:"0" validate:"gte=0,lte=16"`
	// TODO ArtSync?
}

func (d *ArtNet) initialize(base *Device, config map[string]interface{}) (err error) {
	defaults.Set(&d.config)
	err = mapstructure.Decode(&config, &d.config)
	if err != nil {
		return err
	}
	err = validate.Struct(&d.config)
	if err != nil {
		return err
	}
	d.pb, err = newPacketBuilder(base.Config.PixelCount, ArtDMX, byte(d.config.Universe)) // repurpose timeout for universe
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
	service := d.config.IP + ":" + strconv.Itoa(d.config.Port)
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

func (d *ArtNet) getConfig() (c map[string]interface{}) {
	mapstructure.Decode(&d.config, &c)
	return c
}
