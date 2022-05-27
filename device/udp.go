package device

import (
	"ledfx/color"
	"ledfx/logger"
	"net"
	"strconv"
)

type UDP struct {
	Config     UDPConfig
	connection net.Conn
	pb         *packetBuilder
}

type UDPConfig struct {
	NetworkerConfig
	protocol UDPProtocol
	timeout  int
}

func (d *UDP) initialize(base *Device) (err error) {
	d.pb, err = NewPacketBuilder(base.Config.PixelCount, d.Config.protocol, byte(d.Config.timeout))
	return err
}

func (d *UDP) send(p color.Pixels) (err error) {
	_, err = d.connection.Write(d.pb.Build(p))
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
