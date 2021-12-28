package device

import (
	"ledfx/color"
)

type Device interface {
	Init() error
	SendData(colors []color.Color) error
	Close() error
}
