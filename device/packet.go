package device

import (
	"errors"
	"fmt"
	"ledfx/color"
)

var errTooManyPx = errors.New("too many pixels for the packet type")

type packetBuilder struct {
	pixelCount int              // Number of pixels
	protocol   UDPProtocol      // UDP packet type
	timeout    byte             // Number of seconds timeout to include in packet (if protocol allows)
	packets    [][]byte         // Working array for building packet. Might be multiple packets for given pixels
	rgbw       color.PixelsRGBW // Working array for converting to RGBW color space
}

func NewPacketBuilder(pixelCount int, protocol UDPProtocol, timeout byte) (pb *packetBuilder, err error) {
	pb = &packetBuilder{
		pixelCount: pixelCount,
		protocol:   protocol,
		timeout:    timeout,
	}
	switch protocol {
	case WARLS:
		pb.packets = make([][]byte, 1)
		if pixelCount > 255 {
			return pb, errTooManyPx
		}
		pb.packets[0] = make([]byte, 2+pixelCount*4)
		pb.packets[0][0] = protocol.Byte()
		pb.packets[0][1] = timeout
	case DRGB:
		pb.packets = make([][]byte, 1)
		if pixelCount > 490 {
			return pb, errTooManyPx
		}
		pb.packets[0] = make([]byte, 2+pixelCount*3)
		pb.packets[0][0] = protocol.Byte()
		pb.packets[0][1] = timeout
	case DRGBW:
		pb.packets = make([][]byte, 1)
		if pixelCount > 367 {
			return pb, errTooManyPx
		}
		pb.rgbw = make(color.PixelsRGBW, pixelCount)
		pb.packets[0] = make([]byte, 2+pixelCount*4)
		pb.packets[0][0] = protocol.Byte()
		pb.packets[0][1] = timeout
	case DNRGB:
		if pixelCount > 65536 {
			return pb, errTooManyPx
		}
		full_packets := pixelCount / 489
		remainder := pixelCount % 489
		pb.packets = make([][]byte, full_packets+1)

		for i := 0; i < full_packets; i++ {
			pb.packets[i] = make([]byte, 4+489*3)
		}
		pb.packets[full_packets] = make([]byte, 4+remainder*3)

		for i := 0; i <= full_packets; i++ {
			start := uint16(i * 489)
			pb.packets[i][0] = protocol.Byte()
			pb.packets[i][1] = timeout
			pb.packets[i][2] = byte(start >> 8)
			pb.packets[i][3] = byte(start)
		}
	case DDP:
		if pixelCount > 480*16 {
			return pb, errTooManyPx
		}
		full_packets := pixelCount / 480
		remainder := pixelCount % 480
		dlen := 480
		pb.packets = make([][]byte, full_packets+1)

		// constructs the headers for the packets we'll be sending
		// see: http://www.3waylabs.com/ddp/
		for i := 0; i <= full_packets; i++ {
			pb.packets[i] = make([]byte, 10+480*3)
			offset := i * 480
			if i < full_packets {
				pb.packets[i][0] = DDP_HEADER
			} else {
				pb.packets[i][0] = DDP_HEADER | DDP_PUSH
				dlen = remainder
			}
			pb.packets[i][1] = uint8(i)
			pb.packets[i][2] = DDP_DTYPE
			pb.packets[i][3] = DDP_DEST
			pb.packets[i][4] = byte(offset >> 24)
			pb.packets[i][5] = byte(offset >> 16)
			pb.packets[i][6] = byte(offset >> 8)
			pb.packets[i][7] = byte(offset)
			pb.packets[i][8] = byte(dlen >> 8)
			pb.packets[i][9] = byte(dlen)
		}
	default:
		return pb, fmt.Errorf("unknown protocol: %s", protocol)
	}
	return pb, nil
}

func (pb *packetBuilder) Build(p color.Pixels) {
	switch pb.protocol {
	case WARLS: // not the smallest packet possible, but who uses warls. plus, most effects change lots of pixels anyway.
		for i, c := range p {
			pb.packets[0][i*4+2] = byte(i)
			pb.packets[0][i*4+3] = byte(c[0] * 255)
			pb.packets[0][i*4+4] = byte(c[1] * 255)
			pb.packets[0][i*4+5] = byte(c[2] * 255)
		}
	case DRGB:
		for i, c := range p {
			pb.packets[0][i*3+2] = byte(c[0] * 255)
			pb.packets[0][i*3+3] = byte(c[1] * 255)
			pb.packets[0][i*3+4] = byte(c[2] * 255)
		}
	case DRGBW:
		p.ToRGBW(pb.rgbw)
		for i, c := range pb.rgbw {
			pb.packets[0][i*4+2] = byte(c[0] * 255)
			pb.packets[0][i*4+3] = byte(c[1] * 255)
			pb.packets[0][i*4+4] = byte(c[2] * 255)
			pb.packets[0][i*4+5] = byte(c[2] * 255)
		}
	case DNRGB:
		for i, c := range p {
			j := i / 489
			k := i % 489
			pb.packets[j][k*3+4] = byte(c[0] * 255)
			pb.packets[j][k*3+5] = byte(c[1] * 255)
			pb.packets[j][k*3+6] = byte(c[2] * 255)
		}
	case DDP:
		for i, c := range p {
			j := i / 480
			k := i % 480
			pb.packets[j][k*3+10] = byte(c[0] * 255)
			pb.packets[j][k*3+10] = byte(c[1] * 255)
			pb.packets[j][k*3+10] = byte(c[2] * 255)
		}
	}
}
