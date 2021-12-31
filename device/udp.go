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
	warls = 0x01
	drgb  = 0x02
	drgbw = 0x03
	dnrgb = 0x04
	// TODO: adalight
	// TODO: openrgb
)

// UDPProtocols maps the string representation of the UDP packet type to the byte representation
var UDPProtocols = map[string]byte{
	"WARLS": warls,
	"DRGB":  drgb,
	"DRGBW": drgbw,
	"DNRGB": dnrgb,
}

// UDPDevice is a device that uses UDP to send data to a WLED device
type UDPDevice struct {
	Connection net.Conn
	Config     config.DeviceConfig
}

// enforce the Device interface
var _ Device = (*UDPDevice)(nil)

// ColorsToBytes flattens the array of colors and converts them to bytes
func ColorsToBytes(colors []color.Color, includeWhite bool) []byte {
	var multiplier int
	if includeWhite {
		multiplier = 4
	} else {
		multiplier = 3
	}
	bytes := make([]byte, len(colors)*multiplier)
	for i, c := range colors {
		bytes[i*multiplier] = byte(c[0] * 255)
		bytes[i*multiplier+1] = byte(c[1] * 255)
		bytes[i*multiplier+2] = byte(c[2] * 255)
		if includeWhite {
			// currently unused white channel
			bytes[i*multiplier+3] = byte(0x00)
		}
	}
	return bytes
}

// Init initializes the UDP device
func (d *UDPDevice) Init() error {
	hostName := d.Config.IpAddress

	service := hostName + ":" + strconv.Itoa(d.Config.Port)

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

// SendData sends the data to the UDP device over UDP
func (d *UDPDevice) SendData(colors []color.Color, ledOffset int) error {
	if d.Connection == nil {
		return errors.New("Device must first be initialized")
	}

	packet, err := d.BuildPacket(colors, ledOffset)
	if err != nil {
		return err
	}

	logger.Logger.Debug("Sending Data: ", packet)
	_, err = d.Connection.Write(packet)
	if err != nil {
		return err
	}
	return nil
}

// BuildPacket builds the UDP packet to send to the device
func (d *UDPDevice) BuildPacket(colors []color.Color, ledOffset int) ([]byte, error) {
	if d.Config.UdpPacketType == "WARLS" {
		if len(colors) > 255 {
			return nil, &InvalidLedCountError{Count: len(colors), Min: 0, Max: 255}
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
	if protocol == 0x00 {
		protocol = dnrgb // default to DNRGB https://github.com/Aircoookie/WLED/wiki/UDP-Realtime-Control
	}
	protocol = byte(protocol)
	var timeout byte = 0x01
	var offset []byte
	if protocol == warls {
		offset = []byte{byte(ledOffset)}
	} else if protocol == dnrgb {
		offset = []byte{byte(ledOffset), byte(ledOffset >> 8)}
	}
	packet := []byte{protocol, timeout}

	packet = append(packet, offset...)

	data := ColorsToBytes(colors, protocol == drgbw)
	packet = append(packet, data...)
	return packet, nil
}
