package device

import (
	"errors"
	"fmt"
	"ledfx/color"
	"ledfx/config"
	"log"

	"github.com/Hundemeier/go-sacn/sacn"
	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

var transmitter *sacn.Transmitter

var cid = [16]byte{0x6C, 0x65, 0x64, 0x66, 0x78, 0x20, 0x3A, 0x29} // sender ID, "ledfx :)"

type E131 struct {
	Config     E131Config
	chs        []chan<- [512]byte
	pixelCount int
}

type E131Config struct {
	IPs       []string `mapstructure:"ips" json:"ips" description:"Unicast IP addresses on the LAN" validate:"required_if=Multicast false,omitempty,dive,ip"`
	Port      int      `mapstructure:"port" json:"port" description:"Port number the E1.31 device is listening on" default:"5568" validate:"gte=0,lte=65535"`
	Universe  int      `mapstructure:"universe" json:"universe" description:"Starting universe for DMX data. 170 pixels per universe." default:"0" validate:"gte=0,lte=65535"`
	Multicast bool     `mapstructure:"multicast" json:"multicast" description:"Broadcast data via multicast UDP" default:"false" validate:""`
	//
}

func (d *E131) initialize(base *Device, c map[string]interface{}) (err error) {
	defaults.Set(&d.Config)
	err = mapstructure.Decode(&c, &d.Config)
	if err != nil {
		return err
	}
	err = validate.Struct(&d.Config)
	if err != nil {
		return err
	}
	// make sure we dont have too many pixels (i'd be amazed if this limit was approached...)
	d.pixelCount = base.Config.PixelCount
	if d.pixelCount > 170*65535 {
		return errTooManyPx
	}
	// create our transmitter if there isn't one already
	if transmitter != nil {
		return
	}
	settings := config.GetSettings()
	hostport := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	t, err := sacn.NewTransmitter(hostport, cid, "transmitter")
	if err != nil {
		transmitter = &t
	}
	return err
}

func (d *E131) send(p color.Pixels) (err error) {
	for i, c := range p {
		j := i / 170
		k := i % 170
		data := [512]byte{}
		data[k*3+0] = byte(c[0] * 255)
		data[k*3+1] = byte(c[1] * 255)
		data[k*3+2] = byte(c[2] * 255)
		if k == 169 {
			d.chs[j] <- data
		}
	}
	return nil
}

func (d *E131) connect() (err error) {
	// calculate how many universes we need
	uniStart := uint16(d.Config.Universe)
	uniCount := uint16(d.pixelCount / 170)
	if d.pixelCount%170 > 0 {
		uniCount++
	}
	d.chs = make([]chan<- [512]byte, uniCount)
	// make sure our universes arent overlapping with other active universes
	for i := uint16(0); i < uniCount; i++ {
		if transmitter.IsActivated(i + uniStart) {
			d.disconnect()
			return fmt.Errorf("e1.31 universe %d is already in use", i)
		}
	}
	// activate our universes
	for i := uint16(0); i < uniCount; i++ {
		d.chs[i], err = transmitter.Activate(i + uniStart)
		if err != nil {
			return err
		}
		transmitter.SetMulticast(i+uniStart, d.Config.Multicast)
		errs := transmitter.SetDestinations(i+uniStart, d.Config.IPs)
		if len(errs) != 0 {
			d.disconnect()
			return errors.New("invalid unicast IP address")
		}
	}
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (d *E131) disconnect() error {
	for _, ch := range d.chs {
		close(ch)
	}
	return nil
}
