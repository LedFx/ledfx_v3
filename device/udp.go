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
	Config     UDPConfig
	connection net.Conn
	pb         *packetBuilder
}

type UDPConfig struct {
	NetworkerConfig
	Protocol string `mapstructure:"protocol" json:"protocol" description:"UDP packet type" default:"DRGB" validate:"oneof=WARLS DRGB DRGBW DNRGB DDP"`
	Timeout  int    `mapstructure:"timeout" json:"timeout" description:"How long between it takes the device to return to normal state after LedFx stops sending data to it" default:"2" validate:"gte=0,lte=255"`
}

func (d *UDP) initialize(base *Device, config map[string]interface{}) (err error) {
	//d.Config = config.(UDPConfig)
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
	protocol := UDPProtocol(d.Config.Protocol)
	d.pb, err = NewPacketBuilder(base.Config.PixelCount, protocol, byte(d.Config.Timeout))
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

func (d *UDP) disconnect() error {
	return d.connection.Close()
}
