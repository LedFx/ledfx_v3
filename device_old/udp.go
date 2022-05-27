package device_old

import (
	"errors"
	"ledfx/color"
	"ledfx/config"
	"ledfx/logger"
	"net"
)

const (
	WARLS = 0x01
	DRGB  = 0x02
	DRGBW = 0x03
	DNRGB = 0x04
	// TODO: Add more here
)

var UDPProtocols = map[string]byte{
	"WARLS": WARLS,
	"DRGB":  DRGB,
	"DRGBW": DRGBW,
	"DNRGB": DNRGB,
}

type UDPDevice struct {
	Name       string
	Port       int
	Connection net.Conn
	Protocol   byte
	Config     config.DeviceConfig
	pb         *PacketBuilder
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
func (d *UDPDevice) Init() error {
	// hostName := d.Config.IpAddress

	// service := hostName + ":" + strconv.Itoa(d.Port)
	service := d.Config.IpAddress + ":21324"

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

func (d *UDPDevice) Close() error {
	err := d.Connection.Close()
	if err != nil {
		return err
	}
	return nil
}

func (d *UDPDevice) SendData(colors []color.Color, timeout byte) error {
	if d.Connection == nil {
		return errors.New("device must first be initialized")
	}

	packet := d.BuildPacket(colors, timeout)

	// logger.Logger.Debug("Sending Data: ", packet)
	_, err := d.Connection.Write(packet)
	if err != nil {
		return err
	}
	return nil
}

func (d *UDPDevice) BuildPacket(colors []color.Color, timeout byte) []byte {
	// TODO: read from config
	var protocol byte
	if d.Protocol == 0x00 {
		// use default protocol
		d.Protocol = 0x04 // DNRGB https://github.com/Aircoookie/WLED/wiki/UDP-Realtime-Control
	}
	protocol = byte(d.Protocol)
	// TODO: read from config
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

func (d *UDPDevice) PacketBuilder() *PacketBuilder {
	return d.pb
}
