package device

import (
	"errors"
	"fmt"
	"ledfx/color"
	"ledfx/logger"
	"log"
	"runtime"

	"github.com/Hundemeier/go-sacn/sacn"
	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

var cid = [16]byte{0x6C, 0x65, 0x64, 0x66, 0x78, 0x20, 0x3A, 0x29} // sender ID, "ledfx :)"
var transmitter sacn.Transmitter
var transmitter_init bool = false

func init() {
	if transmitter_init {
		return
	}
	var err error
	transmitter, err = sacn.NewTransmitter("", cid, "LedFx")
	transmitter_init = true
	if err != nil {
		log.Fatal("Failed to initialise E1.31 transmitter")
	}
}

type E131 struct {
	config     E131Config
	chs        []chan<- [512]byte
	pixelCount int
}

type E131Config struct {
	IPs       []string `mapstructure:"ips" json:"ips" description:"Unicast IP addresses on the LAN" validate:"required_if=Multicast false,omitempty,dive,ip"`
	Port      int      `mapstructure:"port" json:"port" description:"Port number the E1.31 device is listening on" default:"5568" validate:"gte=0,lte=65535"`
	Universe  int      `mapstructure:"universe" json:"universe" description:"Starting universe for DMX data. 170 pixels per universe." default:"1" validate:"gte=1,lte=65535"`
	Multicast bool     `mapstructure:"multicast" json:"multicast" description:"Broadcast data via multicast UDP" default:"false" validate:""`
}

func (d *E131) initialize(base *Device, c map[string]interface{}) (err error) {
	defaults.Set(&d.config)
	err = mapstructure.Decode(&c, &d.config)
	if err != nil {
		return err
	}
	err = validate.Struct(&d.config)
	if err != nil {
		return err
	}
	// make sure we dont have too many pixels (i'd be amazed if this limit was approached...)
	d.pixelCount = base.Config.PixelCount
	if d.pixelCount > 170*65535 {
		return errTooManyPx
	}
	return nil
}

func (d *E131) send(p color.Pixels) (err error) {
	data := [512]byte{}
	var j, k int
	for i, c := range p {
		if k == 169 && i%170 == 0 { // if we've looped, send the packet and reset
			d.chs[j] <- data
			data = [512]byte{}
		}
		j = i / 170
		k = i % 170
		data[k*3+0] = byte(c[0] * 255)
		data[k*3+1] = byte(c[1] * 255)
		data[k*3+2] = byte(c[2] * 255)
	}
	// if there's any remainder send the final packet
	if len(p)%170 != 0 {
		d.chs[len(p)/170] <- data
	}
	return nil
}

func (d *E131) connect() (err error) {
	// calculate how many universes we need
	uniStart := uint16(d.config.Universe)
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
		if runtime.GOOS == "windows" && d.config.Multicast {
			logger.Logger.WithField("context", "E1.31 sACN").Error("Multicast not supported on Windows")
		} else {
			transmitter.SetMulticast(i+uniStart, d.config.Multicast)
		}
		errs := transmitter.SetDestinations(i+uniStart, d.config.IPs)
		if len(errs) != 0 {
			d.disconnect()
			return errors.New("invalid unicast IP address")
		}
	}
	return err
}

func (d *E131) disconnect() error {
	for _, ch := range d.chs {
		close(ch)
	}
	return nil
}

func (d *E131) getConfig() (c map[string]interface{}) {
	mapstructure.Decode(&d.config, &c)
	return c
}
