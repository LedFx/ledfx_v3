/*
Package color provides basic tools for interpreting colors for LedFX.
LedFx colors internally are all [3]float64 with values 0-1.
These can represent HSV or RGB color space.
Only at the final step of effect processing, before pixels
are sent to the device, are they multiplied up to 256.
*/
package color

import (
	"errors"
	"strconv"
	"strings"
)

type Color [3]float64
type ColorRGBW [4]float64
type Pixels []Color
type PixelsRGBW []ColorRGBW

var Full Color = Color{1, 1, 1}

var errInvalidColor = errors.New("invalid color")

// NewColor Parses string to ledfx color. "#ff00ff" / "rgb(255,0,255)" / "red" -> [1., 0., 1.]
func NewColor(color_string string) (col Color, err error) {
	color_string = strings.ToLower(color_string)
	predef, isPredef := LedFxColors[color_string]
	switch {
	case isPredef: // Color is predefined
		col, err = parseHex(predef)
	case strings.HasPrefix(color_string, "rgb("): // "rgb(0, 127, 255)"
		col, err = parseRGB(color_string)
	case strings.HasPrefix(color_string, "#"): // "#0088ff"
		col, err = parseHex(color_string)
	default:
		return col, errInvalidColor
	}
	if err != nil {
		return Color{}, err
	}
	return col, err
}

func parseRGB(rgb_color_string string) (col Color, err error) {
	rgb_color_string = strings.ReplaceAll(rgb_color_string, " ", "")
	rgb_color_string = strings.TrimLeft(rgb_color_string, "rgb(")
	rgb_color_string = strings.TrimRight(rgb_color_string, ")")
	for sub_color, val := range strings.Split(rgb_color_string, ",") {
		col[sub_color], err = strconv.ParseFloat(val, 64)
		col[sub_color] /= 255
		if col[sub_color] < 0 || col[sub_color] > 1 {
			err = errInvalidColor
		}
	}
	return col, err
}

func parseHex(hex_color_string string) (col Color, err error) {
	if len(hex_color_string) != 7 {
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
	col[0] = float64(hexToByte(hex_color_string[1])<<4+hexToByte(hex_color_string[2])) / 255
	col[1] = float64(hexToByte(hex_color_string[3])<<4+hexToByte(hex_color_string[4])) / 255
	col[2] = float64(hexToByte(hex_color_string[5])<<4+hexToByte(hex_color_string[6])) / 255
	return col, err
}
