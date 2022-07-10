package device

import (
	"ledfx/color"
	"ledfx/logger"
	"runtime"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
	"go.bug.st/serial"
)

type Serial struct {
	config SerialConfig
	port   serial.Port
	pb     *packetBuilder
}

type SerialConfig struct {
	BaudRate int    `mapstructure:"baudrate" json:"baudrate" description:"Communication speed" default:"115200" validate:"oneof=115200 230400 460800 500000 576000 921600 1000000 1500000"`
	Port     string `mapstructure:"port" json:"port" description:"USB port" validate:"required"`
	Protocol string `mapstructure:"protocol" json:"protocol" description:"Serial packet type" default:"Adalight" validate:"oneof=TPM2 Adalight"`
}

func (s *Serial) initialize(base *Device, config map[string]interface{}) error {
	defaults.Set(&s.config)
	err := mapstructure.Decode(&config, &s.config)
	if err != nil {
		return err
	}
	err = validate.Struct(&s.config)
	if err != nil {
		return err
	}
	protocol := Protocol(s.config.Protocol)
	s.pb, err = newPacketBuilder(base.Config.PixelCount, protocol, byte(0))
	return err
}

func (s *Serial) send(p color.Pixels) error {
	s.pb.Build(p)
	var err error
	for i := range s.pb.packets {
		_, err = s.port.Write(s.pb.packets[i])
	}
	return err
}

func (s *Serial) connect() error {
	mode := &serial.Mode{
		BaudRate: s.config.BaudRate,
	}
	var err error
	s.port, err = serial.Open(s.config.Port, mode)
	if e, ok := err.(*serial.PortError); ok {
		if e.Code() == serial.PermissionDenied && runtime.GOOS == "linux" {
			logger.Logger.WithField("context", "Serial").Error("Try adding your user to 'dialout' group - https://askubuntu.com/q/210177")
		}
	}
	return err
}

func (s *Serial) disconnect() error {
	return s.port.Close()
}

func (s *Serial) getConfig() (c map[string]interface{}) {
	mapstructure.Decode(&s.config, &c)
	return c
}
