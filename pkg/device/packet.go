package device

import (
	"errors"
	"fmt"

	"github.com/LedFx/ledfx/pkg/color"
)

var errTooManyPx = errors.New("too many pixels for the packet type")

type packetBuilder struct {
	pixelCount int              // Number of pixels
	protocol   Protocol         // Packet type
	timeout    byte             // Number of seconds timeout to include in packet (if protocol allows)
	packets    [][]byte         // Working array for building packet. Might be multiple packets for given pixels
	rgbw       color.PixelsRGBW // Working array for converting to RGBW color space
}

func newPacketBuilder(pixelCount int, protocol Protocol, timeout byte) (pb *packetBuilder, err error) {
	pb = &packetBuilder{
		pixelCount: pixelCount,
		protocol:   protocol,
		timeout:    timeout,
	}
	// make the packet headers in advance, so they're made just once
	switch protocol {
	case WARLS:
		pb.packets = make([][]byte, 1)
		if pixelCount > 255 {
			return pb, errTooManyPx
		}
		pb.packets[0] = make([]byte, 2+pixelCount*4)
		pb.packets[0][0] = byte(1)
		pb.packets[0][1] = timeout
	case DRGB:
		pb.packets = make([][]byte, 1)
		if pixelCount > 490 {
			return pb, errTooManyPx
		}
		pb.packets[0] = make([]byte, 2+pixelCount*3)
		pb.packets[0][0] = byte(2)
		pb.packets[0][1] = timeout
	case DRGBW:
		pb.packets = make([][]byte, 1)
		if pixelCount > 367 {
			return pb, errTooManyPx
		}
		pb.rgbw = make(color.PixelsRGBW, pixelCount)
		pb.packets[0] = make([]byte, 2+pixelCount*4)
		pb.packets[0][0] = byte(3)
		pb.packets[0][1] = timeout
	case DNRGB:
		if pixelCount > 65535 {
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
			pb.packets[i][0] = byte(4)
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
	case ADA:
		if pixelCount > 4096 { // soft max, could remove
			return pb, errTooManyPx
		}
		pc := uint16(pixelCount)
		pb.packets = make([][]byte, 1)
		pb.packets[0] = make([]byte, 6+pixelCount*3)
		pb.packets[0][0] = 'A'
		pb.packets[0][1] = 'd'
		pb.packets[0][2] = 'a'
		pb.packets[0][3] = byte(pc >> 8)
		pb.packets[0][4] = byte(pc)
		pb.packets[0][5] = byte(pc>>8) ^ byte(pc) ^ 0x55
	case TPM2:
		if pixelCount > 21845 {
			return pb, errTooManyPx
		}
		pc := uint16(pixelCount)
		pb.packets = make([][]byte, 1)
		pb.packets[0] = make([]byte, 5+pixelCount*3)
		pb.packets[0][0] = 0xC9
		pb.packets[0][1] = 0xDA
		pb.packets[0][2] = byte(pc >> 8)
		pb.packets[0][3] = byte(pc)
		pb.packets[0][4+pixelCount*3] = 0x36
	case ArtDMX:
		universe := timeout // repurpose timeout as universe
		if pixelCount > (int(255-universe) * 170) {
			return pb, errTooManyPx
		}
		full_packets := pixelCount / 170
		remainder := pixelCount % 170
		pb.packets = make([][]byte, full_packets+1)

		for i := 0; i <= full_packets; i++ {
			var dlen uint16 = 510
			if i == full_packets {
				pb.packets[i] = make([]byte, 18+remainder*3)
				dlen = uint16(remainder * 3)
			} else {
				pb.packets[i] = make([]byte, 18+170*3)
			}
			opCode := 0x5000
			ver := uint16(14)
			pb.packets[i][0] = 'A'
			pb.packets[i][1] = 'r'
			pb.packets[i][2] = 't'
			pb.packets[i][3] = '-'
			pb.packets[i][4] = 'N'
			pb.packets[i][5] = 'e'
			pb.packets[i][6] = 't'
			pb.packets[i][7] = 0x00
			pb.packets[i][8] = byte(opCode)
			pb.packets[i][9] = byte(opCode >> 8)
			pb.packets[i][10] = byte(ver >> 8)
			pb.packets[i][11] = byte(ver)
			pb.packets[i][12] = 0x00               // seq, would be implemented with sync.
			pb.packets[i][13] = 0x00               // physical
			pb.packets[i][14] = universe + byte(i) // sub uni
			pb.packets[i][15] = 0x00               // net (not used?)
			pb.packets[i][16] = byte(dlen >> 8)
			pb.packets[i][17] = byte(dlen)
		}

	default:
		return pb, fmt.Errorf("unknown protocol: %s", protocol)
	}
	return pb, nil
}

func (pb *packetBuilder) Build(p color.Pixels) {
	switch pb.protocol {
	// update the led data in the packets
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
			pb.packets[j][k*3+11] = byte(c[1] * 255)
			pb.packets[j][k*3+12] = byte(c[2] * 255)
		}
	case ADA:
		for i, c := range p {
			pb.packets[0][i*3+6] = byte(c[0] * 255)
			pb.packets[0][i*3+7] = byte(c[1] * 255)
			pb.packets[0][i*3+8] = byte(c[2] * 255)
		}
	case TPM2:
		for i, c := range p {
			pb.packets[0][i*3+4] = byte(c[0] * 255)
			pb.packets[0][i*3+5] = byte(c[1] * 255)
			pb.packets[0][i*3+6] = byte(c[2] * 255)
		}
	case ArtDMX:
		for i, c := range p {
			j := i / 170
			k := i % 170
			pb.packets[j][k*3+18] = byte(c[0] * 255)
			pb.packets[j][k*3+19] = byte(c[1] * 255)
			pb.packets[j][k*3+20] = byte(c[2] * 255)
		}
	}
}
