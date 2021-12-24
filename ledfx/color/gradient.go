// Package color provides basic tools for interpreting colors for LedFX
package color

import (
	"errors"
	"strconv"
	"strings"
)

var errInvalidGradient = errors.New("invalid gradient")

type gradient struct {
	mode      string
	angle     int64
	colors    [][3]float64
	positions []float64
}

/*
Parses gradient from string of format eg.
"linear-gradient(90deg, rgb(100, 0, 255) 0%, #800000 50%, #ec77ab 100%)"
where each color is associated with a % value for its position in the gradient
each color can be hex or rgb format
*/
func NewGradient(gs string) (g gradient, err error) {
	var splits []string
	gs = strings.ToLower(gs)
	gs = strings.Replace(gs, " ", "", -1)
	splits = strings.SplitN(gs, "(", 2)
	mode := splits[0]
	mode = strings.TrimSuffix(mode, "-gradient")
	angleColorPos := splits[1]
	angleColorPos = strings.TrimRight(angleColorPos, ")")
	splits = strings.SplitN(angleColorPos, ",", 2)
	if (len(splits) != 2) || !strings.HasSuffix(splits[0], "deg") {
		err = errInvalidGradient
		return gradient{}, err
	}
	angleStr := splits[0]
	angleStr = strings.TrimSuffix(angleStr, "deg")
	angle, err := strconv.ParseInt(angleStr, 10, 64)
	colorPos := splits[1]
	splits = strings.SplitAfter(colorPos, "%,")

	var colors = make([][3]float64, len(splits))
	var positions = make([]float64, len(splits))
	var cp_split []string
	var c [3]float64
	var p float64
	for i, cp := range splits {
		cp = strings.TrimRight(cp, "%,")
		switch cp[0:1] {
		case "r": // rgb style
			cp_split = strings.SplitAfter(cp, ")")
			c, err = NewColor(cp_split[0])
			if err != nil {
				break
			}
			p, err = strconv.ParseFloat(cp_split[1], 64)
			p /= 100
		case "#": // hex style
			c, err = NewColor(cp[0:7])
			if err != nil {
				break
			}
			p, err = strconv.ParseFloat(cp[7:], 64)
		default:
			err = errInvalidGradient
		}
		if err != nil {
			break
		}
		colors[i] = c
		positions[i] = p
	}
	if err != nil {
		err = errInvalidGradient
		return gradient{}, err
	}
	g = gradient{
		mode:      mode,
		angle:     angle,
		colors:    colors,
		positions: positions,
	}
	return g, err
}
