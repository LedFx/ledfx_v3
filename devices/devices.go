package devices

import "errors"

const (
	udp = iota
	e131
	// TODO: Add more here
)

type Device struct {
	Name string
	Type int
	// TODO: Add more here
}

func SendDeviceData(device Device, data []byte) error {
	if device.Type == udp {
		udpDevice := UdpDevice{
			Device: device,
			Port:   19446, // TODO: Get this from the config
		}
		SendUdpData(udpDevice, data)
	} else if device.Type == e131 {
		e131Device := E131Device{
			Device: device,
		}
		SendE131Data(e131Device, data)
	} else {
		return errors.New("Unknown device type")
	}
	return nil
}
