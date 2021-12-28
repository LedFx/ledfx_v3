package device

import (
	"errors"
	"ledfx/color"
	"ledfx/config"
	"ledfx/logger"
	"net"
	"strconv"
)

const (
	WARLS = 0x01
	DRGB  = 0x02
	DRGBW = 0x03
	DNRGB = 0x04
	// TODO: Add more here
)

var UdpProtocols = map[string]byte{
	"WARLS": WARLS,
	"DRGB":  DRGB,
	"DRGBW": DRGBW,
	"DNRGB": DNRGB,
}

type UdpDevice struct {
	Name       string
	Port       int
	Connection net.Conn
	Protocol   byte
	Config     config.DeviceConfig
}

// Flatten array and convert to bytes
func ColorsToBytes(colors []color.Color) []byte {
	bytes := make([]byte, len(colors)*3)
	for i, c := range colors {
		bytes[i*3] = byte(c[0] * 255)
		bytes[i*3+1] = byte(c[1] * 255)
		bytes[i*3+2] = byte(c[2] * 255)
	}
	return bytes
}

// Need to store the connection on the device struct
func (d *UdpDevice) Init() error {
	hostName := d.Config.IpAddress

	service := hostName + ":" + strconv.Itoa(d.Port)

	RemoteAddr, err := net.ResolveUDPAddr("udp", service)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, RemoteAddr)

	if err != nil {
		return err
	}

	d.Connection = conn

	logger.Logger.Debugf("Established connection to %s \n", service)
	logger.Logger.Debugf("Remote UDP address : %s \n", conn.RemoteAddr().String())
	logger.Logger.Debugf("Local UDP client address : %s \n", conn.LocalAddr().String())
	return nil
}

func (d *UdpDevice) Close() error {
	err := d.Connection.Close()
	if err != nil {
		return err
	}
	return nil
}

func (d *UdpDevice) SendData(colors []color.Color) error {
	if d.Connection == nil {
		return errors.New("Device must first be initialized")
	}

	packet := d.BuildPacket(colors)

	logger.Logger.Debug("Sending Data: ", packet)
	_, err := d.Connection.Write(packet)
	if err != nil {
		return err
	}
	return nil
}

func (d *UdpDevice) BuildPacket(colors []color.Color) []byte {
	// TODO: read from config
	var protocol byte
	if d.Protocol == 0x00 {
		// use default protocol
		d.Protocol = 0x04 // DNRGB https://github.com/Aircoookie/WLED/wiki/UDP-Realtime-Control
	}
	protocol = byte(d.Protocol)
	// TODO: read from config
	var timeout byte = 0x01
	// TODO: get from params
	ledOffset := []byte{}
	if d.Protocol == WARLS {
		ledOffset = []byte{0x00}
	} else if d.Protocol == DNRGB {
		ledOffset = []byte{0x00, 0x00}
	}
	packet := []byte{protocol, timeout}

	packet = append(packet, ledOffset...)

	data := ColorsToBytes(colors)
	packet = append(packet, data...)
	return packet
}
