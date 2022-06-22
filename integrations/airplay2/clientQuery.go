package airplay2

import (
	"fmt"
	log "ledfx/logger"
	"net"
	"regexp"

	"github.com/grantmd/go-airplay" //nolint:typecheck
)

func queryDevice(params ClientDiscoveryParameters) (*airplay.AirplayDevice, error) {
	switch {
	case params.DeviceIP != "":
		ip := net.ParseIP(params.DeviceIP)
		if ip == nil {
			return nil, fmt.Errorf("could not parse IP address '%s'", params.DeviceIP)
		}
		return queryDeviceByIP(ip)
	case params.DeviceNameRegex != "":
		return queryDeviceByName(params.DeviceNameRegex)
	default:
		return nil, fmt.Errorf("either DeviceNameRegex or DeviceIP must be populated in the client discovery parameters")
	}
}

func queryDeviceByIP(ip net.IP) (device *airplay.AirplayDevice, err error) {
	ch := make(chan []airplay.AirplayDevice)
	go airplay.Discover(ch)

	for {
		list := <-ch
		for _, dev := range list {
			printDevice(&dev)
			if dev.IP.Equal(ip) {
				return &dev, nil
			}
		}
	}
}

func queryDeviceByName(name string) (device *airplay.AirplayDevice, err error) {
	rxp, err := regexp.Compile(name)
	if err != nil {
		return nil, fmt.Errorf("error compiling regular expression: %w", err)
	}

	ch := make(chan []airplay.AirplayDevice)
	go airplay.Discover(ch)

	for {
		list := <-ch
		for _, dev := range list {
			printDevice(&dev)
			if dev.IP == nil || dev.Type != "airplay" {
				continue
			}

			// Did we find it?
			if rxp.MatchString(dev.Name) {
				// Yes, we did.
				return &dev, nil
			}
		}
	}
}

func printDevice(device *airplay.AirplayDevice) {
	log.Logger.WithField("context", "AirPlay Discovery").Debugf(
		`NAME="%s" SERVER="%s:%d" HOSTNAME="%s" AUDIO="%dch/%dhz/%d-bit" PCM="%v" ALAC="%v"`,
		device.Name,
		device.IP,
		device.Port,
		device.Hostname,
		device.AudioChannels(),
		device.AudioSampleRate(),
		device.AudioSampleSize(),
		determinePCM(device),
		determineALAC(device),
	)
}

func determinePCM(device *airplay.AirplayDevice) bool {
	for _, c := range device.AudioCodecs() {
		if c == 0 {
			return true
		}
	}
	return false
}

func determineALAC(device *airplay.AirplayDevice) bool {
	for _, c := range device.AudioCodecs() {
		if c == 1 {
			return true
		}
	}
	return false
}
