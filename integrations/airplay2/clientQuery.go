package airplay2

import (
	"fmt"
	"github.com/grantmd/go-airplay"
	"net"
	"os"
	"strings"
)

func queryDevice(params ClientDiscoveryParameters) (*airplay.AirplayDevice, error) {
	switch {
	case params.DeviceIP != "":
		ip := net.ParseIP(params.DeviceIP)
		if ip == nil {
			return nil, fmt.Errorf("could not parse IP address '%s'", params.DeviceIP)
		}
		return queryDeviceByIP(ip, false)
	case params.DeviceName != "":
		return queryDeviceByName(params.DeviceName, false)
	default:
		return nil, fmt.Errorf("either DeviceName or DeviceIP must be populated in the client discovery parameters")
	}
}

func queryDeviceByIP(ip net.IP, verbose bool) (device *airplay.AirplayDevice, err error) {
	ch := make(chan []airplay.AirplayDevice)
	go airplay.Discover(ch)

	for {
		list := <-ch
		for _, dev := range list {
			if verbose {
				printDevice(dev)
			}
			if dev.IP.Equal(ip) {
				return &dev, nil
			}
		}
	}
}

func queryDeviceByName(name string, verbose bool) (device *airplay.AirplayDevice, err error) {
	ch := make(chan []airplay.AirplayDevice)
	go airplay.Discover(ch)

	for {
		list := <-ch
		for _, dev := range list {
			if verbose {
				printDevice(dev)
			}
			if dev.IP == nil || dev.Type != "airplay" {
				continue
			}
			// Did we find it?
			if strings.Contains(strings.ToLower(dev.Name), strings.ToLower(name)) {
				// Yes, we did.
				return &dev, nil
			}
		}
	}
}

func printDevice(device airplay.AirplayDevice) {
	_, _ = fmt.Fprintf(os.Stderr, "------BEGIN DEVICE INFO------\n")
	_, _ = fmt.Fprintf(os.Stderr, "%s\n", device.String())
	_, _ = fmt.Fprintf(os.Stderr, "-------END DEVICE INFO-------\n")
}
