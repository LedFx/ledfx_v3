// Package color provides basic tools for interpreting colors for LedFX
package color

import (
	"errors"
	"strconv"
	"strings"
)

var errInvalidGradient = errors.New("invalid gradient")

type Gradient struct {
	mode  string
	angle int64
	// TODO: Do we want this exported or do we want a getter?
	Colors    []Color
	positions []float64
}

/*
Parses gradient from string of format eg.
"linear-gradient(90deg, rgb(100, 0, 255) 0%, #800000 50%, #ec77ab 100%)"
where each color is associated with a % value for its position in the gradient
each color can be hex or rgb format
*/
func NewGradient(gs string) (g Gradient, err error) {
	var splits []string
	gs = strings.ToLower(gs)
	gs = strings.ReplaceAll(gs, " ", "")
	splits = strings.SplitN(gs, "(", 2)
	mode := splits[0]
	mode = strings.TrimSuffix(mode, "-gradient")
	angleColorPos := splits[1]
	angleColorPos = strings.TrimRight(angleColorPos, ")")
	splits = strings.SplitN(angleColorPos, ",", 2)
	if (len(splits) != 2) || !strings.HasSuffix(splits[0], "deg") {
		err = errInvalidGradient
		return Gradient{}, err
	}
	angleStr := splits[0]
	angleStr = strings.TrimSuffix(angleStr, "deg")
	angle, err := strconv.ParseInt(angleStr, 10, 64)
	colorPos := splits[1]
	splits = strings.SplitAfter(colorPos, "%,")

	var colors = make([]Color, len(splits))
	var positions = make([]float64, len(splits))
	var cp_split []string
	var c Color
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
		return Gradient{}, err
	}
	g = Gradient{
		mode:      mode,
		angle:     angle,
		Colors:    colors,
		positions: positions,
	}
	return g, err
}
