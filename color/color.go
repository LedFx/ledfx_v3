// Package color provides basic tools for interpreting colors for LedFX
package color

import (
	"errors"
	"image/color"
	"strconv"
	"strings"
)

/*
LedFx colors internally are all [3]float64 with values 0-1.
Only at the final step of effect processing, before pixels
are sent to the device, are they multiplied up to 256.
*/

type Color [3]float64

var errInvalidColor = errors.New("invalid color")

// NewColor Parses string to ledfx color. "#ff00ff" / "rgb(255,0,255)" / "red" -> [1., 0., 1.]
func NewColor(c string) (col Color, err error) {
	c = strings.ToLower(c)
	predef, isPredef := LedFxColors[c]
	switch {
	case isPredef: // Color is predefined
		col, err = parseHex(predef)
	case strings.HasPrefix(c, "rgb("): // "rgb(0, 127, 255)"
		col, err = parseRGB(c)
	case strings.HasPrefix(c, "#"): // "#0088ff"
		col, err = parseHex(c)
	default:
		return col, errInvalidColor
	}
	return col, nil
}

func parseRGB(c string) (col Color, err error) {
	c = strings.ReplaceAll(c, " ", "")
	c = strings.TrimLeft(c, "rgb(")
	c = strings.TrimRight(c, ")")
	for i, val := range strings.Split(c, ",") {
		col[i], err = strconv.ParseFloat(val, 64)
		col[i] /= 255
		if col[i] < 0 || col[i] > 1 {
			err = errInvalidColor
		}
	}
	return col, err
}

func parseHex(c string) (col Color, err error) {
	if len(c) != 7 {
		err = errInvalidColor
		return col, err
	}
	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		}
		err = errInvalidColor
		return 0
	}
	col[0] = float64(hexToByte(c[1])<<4+hexToByte(c[2])) / 255
	col[1] = float64(hexToByte(c[3])<<4+hexToByte(c[4])) / 255
	col[2] = float64(hexToByte(c[5])<<4+hexToByte(c[6])) / 255
	return col, err
}

func (col Color) NRGBA() color.NRGBA {
	return color.NRGBA{
		R: NormalizeFloat(col[0]),
		G: NormalizeFloat(col[1]),
		B: NormalizeFloat(col[2]),
		A: 255,
	}
}

func NormalizeFloat(f float64) uint8 {
	return uint8(f * 255)
}

func NormalizeColorList(cols []Color) []color.Color {
	nc := make([]color.Color, len(cols))
	for i := range cols {
		nc[i] = cols[i].NRGBA()
	}
	return nc
}

func RandomColor() string {
	for _, v := range LedFxColors {
		return v
	}
	// Return red by default if for some reason the map is empty. It won't be.
	return "#ff0000"
}

func FromBufSliceSum(sum float64) string {
	// sum divided by 37
	return ""
}
