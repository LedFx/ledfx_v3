// Package color provides basic tools for interpreting colors for LedFX
package color

import (
	"errors"
	"strconv"
	"strings"
)

/*
LedFx colors internally are all [3]float64 with values 0-1.
Only at the final step of effect processing, before pixels
are sent to the device, are they multiplied up to 256.
*/

var errInvalid = errors.New("Invalid color")

func ParseString(c string) (col [3]float64, err error) {
	c = strings.ToLower(c)
	switch c[0:1] {
	case "r": // "rgb(0, 127, 255)"
		c = strings.Replace(c, " ", "", -1)
		c = strings.TrimLeft(c, "rgb(")
		c = strings.TrimRight(c, ")")
		for i, val := range strings.Split(c, ",") {
			col[i], err = strconv.ParseFloat(val, 64)
			col[i] /= 255
			if col[i] < 0 || col[i] > 1 {
				err = errInvalid
			}
		}
	case "#": // "#0088ff"
		hexToByte := func(b byte) byte {
			switch {
			case b >= '0' && b <= '9':
				return b - '0'
			case b >= 'a' && b <= 'f':
				return b - 'a' + 10
			}
			err = errInvalid
			return 0
		}

		switch len(c) {
		case 7:
			col[0] = float64(hexToByte(c[1])<<4 + hexToByte(c[2]))
			col[1] = float64(hexToByte(c[3])<<4 + hexToByte(c[4]))
			col[2] = float64(hexToByte(c[5])<<4 + hexToByte(c[6]))
		case 4:
			col[0] = float64(hexToByte(c[1]) * 17)
			col[1] = float64(hexToByte(c[2]) * 17)
			col[2] = float64(hexToByte(c[3]) * 17)
		}
	default:
		err = errInvalid
	}
	return col, err
}
