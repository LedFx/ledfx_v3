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
	Config SerialConfig
	port   serial.Port
	pb     *packetBuilder
}

func baudrateToByte(rate int) byte {
	switch rate {
	case 115200:
		return 0xB0
	case 230400:
		return 0xB1
	case 460800:
		return 0xB2
	case 500000:
		return 0xB3
	case 576000:
		return 0xB4
	case 921600:
		return 0xB5
	case 1000000:
		return 0xB6
	case 1500000:
		return 0xB7
	default:
		return 0xB0
	}
}

type SerialConfig struct {
	BaudRate int    `mapstructure:"baudrate" json:"baudrate" description:"Communication speed" default:"115200" validate:"oneof=115200 230400 460800 500000 576000 921600 1000000 1500000"`
	Port     string `mapstructure:"port" json:"port" description:"USB port" validate:"required"`
	Protocol string `mapstructure:"protocol" json:"protocol" description:"Serial packet type" default:"Adalight" validate:"oneof=TPM2 Adalight"`
}

func (s *Serial) initialize(base *Device, config map[string]interface{}) error {
	defaults.Set(&s.Config)
	err := mapstructure.Decode(&config, &s.Config)
	if err != nil {
		return err
	}
	err = validate.Struct(&s.Config)
	if err != nil {
		return err
	}
	protocol := Protocol(s.Config.Protocol)
	s.pb, err = NewPacketBuilder(base.Config.PixelCount, protocol, byte(0))
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
		BaudRate: s.Config.BaudRate,
	}
	var err error
	s.port, err = serial.Open(s.Config.Port, mode)
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
