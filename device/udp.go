package device

import (
	"ledfx/color"
	"ledfx/logger"
	"net"
	"strconv"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

type UDP struct {
	config     UDPConfig
	connection net.Conn
	pb         *packetBuilder
}

type UDPConfig struct {
	IP       string `mapstructure:"ip" json:"ip" description:"Device IP address on the LAN" validate:"required,ip"`
	Port     int    `mapstructure:"port" json:"port" description:"Port number the device is listening on" default:"21324" validate:"gte=0,lte=65535"`
	Protocol string `mapstructure:"protocol" json:"protocol" description:"UDP packet type" default:"DRGB" validate:"oneof=WARLS DRGB DRGBW DNRGB DDP"`
	Timeout  int    `mapstructure:"timeout" json:"timeout" description:"How many seconds for the device to return to normal state after LedFx stops sending data to it" default:"2" validate:"gte=0,lte=255"`
}

func (d *UDP) initialize(base *Device, config map[string]interface{}) (err error) {
	//d.Config = config.(UDPConfig)
	defaults.Set(&d.config)
	err = mapstructure.Decode(&config, &d.config)
	if err != nil {
		return err
	}
	err = validate.Struct(&d.config)
	if err != nil {
		return err
	}
	protocol := Protocol(d.config.Protocol)
	d.pb, err = newPacketBuilder(base.Config.PixelCount, protocol, byte(d.config.Timeout))
	return err
}

func (d *UDP) send(p color.Pixels) (err error) {
	d.pb.Build(p)
	for i := range d.pb.packets {
		_, err = d.connection.Write(d.pb.packets[i])
	}
	return err
}

func (d *UDP) connect() (err error) {
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

func (d *UDP) disconnect() error {
	return d.connection.Close()
}

func (d *UDP) getConfig() (c map[string]interface{}) {
	mapstructure.Decode(&d.config, &c)
	return c
}
