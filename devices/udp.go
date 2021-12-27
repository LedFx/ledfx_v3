package devices

import (
	"log"
	"net"
)

type UdpDevice struct {
	Port   int
	Device Device
}

func SendUdpPacket() error {
	// TODO: https://ops.tips/blog/udp-client-and-server-in-go/

	hostName := "wled.local"
	portNum := "21324"

	service := hostName + ":" + portNum

	RemoteAddr, err := net.ResolveUDPAddr("udp", service)

	conn, err := net.DialUDP("udp", nil, RemoteAddr)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Established connection to %s \n", service)
	log.Printf("Remote UDP address : %s \n", conn.RemoteAddr().String())
	log.Printf("Local UDP client address : %s \n", conn.LocalAddr().String())

	defer conn.Close()

	// TODO: read from config
	var protocol byte = 0x04 // DNRGB https://github.com/Aircoookie/WLED/wiki/UDP-Realtime-Control
	// write a message to server
	// TODO: read from config
	var timeout byte = 0xff // No timeout
	ledOffset := []byte{0x00, 0x00}
	prefix := []byte{protocol, timeout}
	prefix = append(prefix, ledOffset...)
	message := append(prefix, []byte{0xff, 0xff, 0x00, 0xff, 0xff, 0x00, 0xff, 0xff, 0x00, 0xff, 0xff, 0x00, 0xff, 0xff, 0x00, 0xff, 0xff, 0x00}...)
	log.Println(message)

	_, err = conn.Write(message)
	return nil
}

func SendUdpData(device UdpDevice, data []byte) error {
	return nil
}
