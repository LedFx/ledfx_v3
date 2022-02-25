package device

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
	// TODO: adalight
	// TODO: openrgb
)

// UDPProtocols maps the string representation of the UDP packet type to the byte representation
var UDPProtocols = map[string]byte{
	"WARLS": WARLS,
	"DRGB":  DRGB,
	"DRGBW": DRGBW,
	"DNRGB": DNRGB,
}

// UDPDevice is a device that uses UDP to send data to a WLED device
type UDPDevice struct {
	Connection net.Conn
	Config     config.DeviceConfig
	pb         *PacketBuilder
}

// enforce the Device interface
var _ Device = (*UDPDevice)(nil)

// ColorsToRGBBytes flattens the array of colors and converts them to rgb byte values
func ColorsToRGBBytes(colors []color.Color) []byte {
	bytes := make([]byte, len(colors)*3)
	for i, c := range colors {
		bytes[i*3] = byte(c[0] * 255)
		bytes[i*3+1] = byte(c[1] * 255)
		bytes[i*3+2] = byte(c[2] * 255)
	}
	return bytes
}

// ColorsToRGBWBytes flattens the array of colors and converts them to rgbw byte values
func ColorsToRGBWBytes(colors []color.Color) []byte {
	bytes := make([]byte, len(colors)*4)
	for i, c := range colors {
		bytes[i*4] = byte(c[0] * 255)
		bytes[i*4+1] = byte(c[1] * 255)
		bytes[i*4+2] = byte(c[2] * 255)
		// currently unused white channel
		bytes[i*4+3] = byte(0x00)
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

// Close closes the UDP connection on the device
func (d *UDPDevice) Close() error {
	err := d.Connection.Close()
	if err != nil {
		return err
	}
	return nil
}

func (d *UdpDevice) SendData(colors []color.Color, timeout byte) error {
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

// BuildPacket builds the UDP packet to send to the device
func (d *UdpDevice) BuildPacket(colors []color.Color, timeout byte) []byte {
	if d.Config.UdpPacketType == "WARLS" {
		if len(colors) > 255 {
			return nil, &]{Count: len(colors), Min: 0, Max: 255}
		}
		if ledOffset < 0 || ledOffset > 255 {
			return nil, &InvalidLedOffsetError{Offset: ledOffset, Min: 0, Max: 255}
		}
	} else if d.Config.UdpPacketType == "DRGB" {
		if len(colors) > 490 {
			return nil, &InvalidLedCountError{Count: len(colors), Min: 0, Max: 490}
		}
		if ledOffset != 0 {
			return nil, &InvalidLedOffsetError{Offset: ledOffset, Min: 0, Max: 0}
		}
	} else if d.Config.UdpPacketType == "DRGBW" {
		if len(colors) > 367 {
			return nil, &InvalidLedCountError{Count: len(colors), Min: 0, Max: 367}
		}
		if ledOffset != 0 {
			return nil, &InvalidLedOffsetError{Offset: ledOffset, Min: 0, Max: 0}
		}
	} else if d.Config.UdpPacketType == "DNRGB" {
		if len(colors) > 489 {
			return nil, &InvalidLedCountError{Count: len(colors), Min: 0, Max: 489}
		}
		if ledOffset < 0 || ledOffset > 65535 {
			return nil, &InvalidLedOffsetError{Offset: ledOffset, Min: 0, Max: 65535}
		}
	}

	protocol := UDPProtocols[d.Config.UdpPacketType]
	if d.Protocol == 0x00 {
		// use default protocol
		d.Protocol = DNRGB // DNRGB https://github.com/Aircoookie/WLED/wiki/UDP-Realtime-Control
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

	packet = append(packet, offset...)

	var data []byte
	if protocol == DRGBW {
		data = ColorsToRGBWBytes(colors)
	} else {
		data = ColorsToRGBBytes(colors)
	}
	packet = append(packet, data...)
	return packet, nil
}

func (d *UdpDevice) PacketBuilder() *PacketBuilder {
	return d.pb
}
