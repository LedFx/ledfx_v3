package bluetooth

import (
	"errors"
	"fmt"
	"github.com/muka/go-bluetooth/api"
	"github.com/muka/go-bluetooth/bluez/profile/adapter"
	"github.com/muka/go-bluetooth/bluez/profile/device"
	log "github.com/sirupsen/logrus"
	"ledfx/integrations/bluetooth/util"
	"regexp"
	"sync"
	"time"
)

type Client struct {
	mu sync.Mutex

	dev     *device.Device1
	adapter *adapter.Adapter1

	discoverChan     chan *adapter.DeviceDiscovered
	cancelDiscoverFn func()

	done chan struct{}
}

// NewClient initializes a new Bluetooth adapter client
func NewClient(loglevel log.Level) (cl *Client, err error) {
	cl = &Client{
		mu:   sync.Mutex{},
		done: make(chan struct{}),
	}
	log.SetLevel(loglevel)
	if cl.adapter, err = adapter.GetDefaultAdapter(); err != nil {
		return nil, fmt.Errorf("error getting default Bluetooth adapter: %w", err)
	}
	log.Debugf("Default Bluetooth adapter: %s\n", cl.adapter.Properties.Name)

	if err := cl.adapter.SetPowered(true); err != nil {
		return nil, fmt.Errorf("error powering on Bluetooth adapter: %w", err)
	}

	log.Debugln("Powered on Bluetooth adapter...")

	return cl, nil
}

// SearchAndConnect validates a search criteria (see SearchTargetConfig) and attempts to
// initiate a connection to the requested device once found.
func (cl *Client) SearchAndConnect(config SearchTargetConfig) (err error) {
	var matchFunc func(mac string, name string) (matched bool)

	switch {
	case len(config.DeviceAddress) > 0:
		if config.DeviceAddress, err = util.CleanMacAddress(config.DeviceAddress); err != nil {
			return fmt.Errorf("error cleaning MAC address: %w", err)
		}
		matchFunc = func(mac string, _ string) (matched bool) {
			return mac == config.DeviceAddress
		}
	default:
		if len(config.DeviceRegex) == 0 {
			return fmt.Errorf("either config.DeviceAddress or config.DeviceRegex must be specified")
		}

		rxp, err := regexp.Compile(config.DeviceRegex)
		if err != nil {
			return fmt.Errorf("error compiling regular expression: %w", err)
		}
		matchFunc = func(_ string, name string) (matched bool) {
			return rxp.MatchString(name)
		}
	}

	log.Infof("Starting tryCacheConnect...\n")
	if err := cl.tryCacheConnect(matchFunc, config); err != nil {
		if errors.Is(err, ErrBtDeviceNotFound) {
			go func() {
				log.Infof("Could not find device in cache, starting tryDiscoveryConnect...\n")
				if err := cl.tryDiscoveryConnect(matchFunc, config); err != nil {
					log.Errorf("error attempting connection through discovery: %v\n", err)
				}
			}()
			return nil
		}
		return fmt.Errorf("error attempting connection through device cache: %w", err)
	}
	return nil
}

// WaitConnect waits for the Bluetooth adapter to successfully connect to the device
// requested by SearchAndConnect.
func (cl *Client) WaitConnect() {
	<-cl.done
}

// tryCacheConnect runs matchFunc() on all devices in the adapter cache.
func (cl *Client) tryCacheConnect(matchFunc func(mac string, name string) (matched bool), config SearchTargetConfig) (err error) {
	devices, err := cl.adapter.GetDevices()
	if err != nil {
		return fmt.Errorf("error getting device cache list: %w", err)
	}

	for _, cl.dev = range devices {
		if matchFunc(cl.dev.Properties.Address, cl.dev.Properties.Name) {
			log.Infof("Found requested device in cache: (addr=%s, name=%s)", cl.dev.Properties.Address, cl.dev.Properties.Name)
			break
		}
		log.Debugf("Found non-matching device: (addr=%s, name=%s)", cl.dev.Properties.Address, cl.dev.Properties.Name)
		cl.dev = nil
	}

	if cl.dev != nil {
		go cl.tryConnectForever(config.ConnectRetryCoolDown)
		return nil
	}
	return ErrBtDeviceNotFound
}

// tryDiscoveryConnect runs matchFunc() on all devices discovered by the Bluetooth adapter.
func (cl *Client) tryDiscoveryConnect(matchFunc func(mac string, name string) (matched bool), config SearchTargetConfig) (err error) {
	if cl.discoverChan, cl.cancelDiscoverFn, err = api.Discover(cl.adapter, nil); err != nil {
		return fmt.Errorf("error starting discovery: %w", err)
	}
	defer cl.cancelDiscoverFn()

	for found := range cl.discoverChan {
		// If it's removed, ignore it
		if found.Type == adapter.DeviceRemoved {
			continue
		}

		if cl.dev, err = device.NewDevice1(found.Path); err != nil {
			log.Warnf("Error generating new device from dbus object: %v\n", err)
			continue
		}

		if matchFunc(cl.dev.Properties.Address, cl.dev.Properties.Name) {
			log.Infof("Found requested device: (addr=%s, name=%s)\n", cl.dev.Properties.Address, cl.dev.Properties.Name)
			break
		}
		log.Debugf("Found non-matching device: (addr=%s, name=%s)", cl.dev.Properties.Address, cl.dev.Properties.Name)
		cl.dev = nil
	}

	if cl.dev != nil {
		go cl.tryConnectForever(config.ConnectRetryCoolDown)
		return nil
	}
	return ErrBtDeviceNotFound
}

// tryConnectForever is self-explanatory. It attempts to connect to dev until it succeeds.
func (cl *Client) tryConnectForever(coolDown time.Duration) {
	log.Infof("Attempting to connect to %q indefinitely...\n", cl.dev.Properties.Address)
	for err := cl.dev.Connect(); err != nil; {
		log.Debugf("Error encountered during connection attempt to Bluetooth device: %v (retrying...)\n", err)
		time.Sleep(coolDown)
	}
	log.Infof("Connection to Bluetooth device with address %q succeeded\n", cl.dev.Properties.Name)
	cl.done <- struct{}{}
}
