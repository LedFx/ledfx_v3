package device

import (
	"errors"
	"fmt"
	"ledfx/color"
)

var tooManyPx = errors.New("too many pixels for the packet type")

type packetBuilder struct {
	pixelCount int
	protocol   UDPProtocol
	timeout    byte
	packet     []byte
	rgbw       color.PixelsRGBW
}

func NewPacketBuilder(pixelCount int, protocol UDPProtocol, timeout byte) (pb *packetBuilder, err error) {
	pb = &packetBuilder{
		pixelCount: pixelCount,
		protocol:   protocol,
		timeout:    timeout,
	}
	var packet []byte
	switch protocol {
	case WARLS:
		if pixelCount > 255 {
			return pb, tooManyPx
		}
		packet = make([]byte, 2+pixelCount*4)
	case DRGB:
		if pixelCount > 490 {
			return pb, tooManyPx
		}
		packet = make([]byte, 2+pixelCount*3)
	case DRGBW:
		if pixelCount > 367 {
			return pb, tooManyPx
		}
		packet = make([]byte, 2+pixelCount*4)
		pb.rgbw = make(color.PixelsRGBW, pixelCount)
	default:
		return pb, fmt.Errorf("unknown protocol: %s", protocol)
	}
	packet[0] = byte(protocol)
	packet[1] = timeout
	pb.packet = packet
	return pb, nil
}

func (pb *packetBuilder) Build(p color.Pixels) []byte {
	switch pb.protocol {
	case WARLS: // not the smallest packet possible, but who uses warls. plus, most effects change lots of pixels anyway.
		for i, c := range p {
			pb.packet[i*4+2] = byte(i)
			pb.packet[i*4+3] = byte(c[0] * 255)
			pb.packet[i*4+4] = byte(c[1] * 255)
			pb.packet[i*4+5] = byte(c[2] * 255)
		}
	case DRGB:
		for i, c := range p {
			pb.packet[i*3+2] = byte(c[0] * 255)
			pb.packet[i*3+3] = byte(c[1] * 255)
			pb.packet[i*3+4] = byte(c[2] * 255)
		}
	case DRGBW:
		p.ToRGBW(pb.rgbw)
		for i, c := range pb.rgbw {
			pb.packet[i*4+2] = byte(c[0] * 255)
			pb.packet[i*4+3] = byte(c[1] * 255)
			pb.packet[i*4+4] = byte(c[2] * 255)
			pb.packet[i*4+5] = byte(c[2] * 255)
		}
	}
	return pb.packet
}
