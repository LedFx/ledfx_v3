// Package color provides basic tools for interpreting colors for LedFX
package color

import (
	"errors"
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/muesli/gamut/palette"

	// Side effects
	_ "image/draw"
	_ "image/jpeg"

	_ "golang.org/x/image/riff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vector"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"
)

var errInvalidPalette = errors.New("invalid palette")

// palettes are generated at a fixed size from the palette string.
const PaletteSize int = 300
const paletteSizeFloat float64 = float64(PaletteSize)

type Palette struct {
	mode      string
	angle     int64
	colors    []Color
	positions []float64
	rgb       [PaletteSize]Color // maps the hue value to an RGB color
	rawCSS    string
}

func (p *Palette) Get(pos float64) Color {
	// this makes pos wrap smoothly above 1 and below 0
	// giving the impression of a cyclic color palette
	// see: https://www.desmos.com/calculator/1dhc2nhann
	pos = 1 - math.Abs(math.Mod(math.Abs(pos), 2)-1)
	idx := int(pos*paletteSizeFloat - 1)
	return p.rgb[idx]
}

func (p *Palette) String() string {
	return p.rawCSS
}

func NewPalette(gs string) (g *Palette, err error) {
	predef, isPredef := LedFxPalettes[gs]
	if isPredef {
		return ParsePalette(predef)
	}
	return ParsePalette(gs)
}

/*
Parses palette from string of format eg.
"linear-palette(90deg, rgb(100, 0, 255) 0%, #800000 50%, #ec77ab 100%)"
Each color is associated with a % value for its position in the palette.
Each color can be hex or rgb format
*/
func ParsePalette(gs string) (g *Palette, err error) {
	g = &Palette{
		rawCSS: gs,
	}

	var splits []string
	gs = strings.ToLower(gs)
	gs = strings.ReplaceAll(gs, " ", "")
	splits = strings.SplitN(gs, "(", 2)
	mode := splits[0]
	g.mode = strings.TrimSuffix(mode, "-palette")
	angleColorPos := splits[1]
	angleColorPos = strings.TrimRight(angleColorPos, ")")
	splits = strings.SplitN(angleColorPos, ",", 2)
	if (len(splits) != 2) || !strings.HasSuffix(splits[0], "deg") {
		return nil, errInvalidPalette
	}
	angleStr := splits[0]
	angleStr = strings.TrimSuffix(angleStr, "deg")
	if g.angle, err = strconv.ParseInt(angleStr, 10, 64); err != nil {
		return nil, fmt.Errorf("error parsing angle string: %w", err)
	}
	colorPos := splits[1]
	splits = strings.SplitAfter(colorPos, "%,")

	g.colors = make([]Color, len(splits))
	g.positions = make([]float64, len(splits))
	var cpSplit []string
	var c Color
	var p float64
	for i, cp := range splits {
		cp = strings.TrimRight(cp, "%,")
		switch cp[0:1] {
		case "r": // rgb style
			cpSplit = strings.SplitAfter(cp, ")")
			c, err = NewColor(cpSplit[0])
			if err != nil {
				break
			}
			p, err = strconv.ParseFloat(cpSplit[1], 64)
			p /= 100
		case "#": // hex style
			c, err = NewColor(cp[0:7])
			if err != nil {
				break
			}
			p, err = strconv.ParseFloat(cp[7:], 64)
		default:
			err = errInvalidPalette
		}
		if err != nil {
			break
		}
		g.colors[i] = c
		g.positions[i] = p
	}
	if err != nil {
		return nil, errInvalidPalette
	}
	if (g.positions[0] != 0) || (g.positions[len(g.positions)-1] != 1) {
		return nil, errors.New("palette color positions must start at 0% and end at 100%")
	}

	// Create the RGB color array

	return g, err
}

// Creates smooth color changes.
// See: https://www.desmos.com/calculator/uh2s7dhmkw
func ease(chunk_len int, start_val, end_val, slope float64) []float64 {
	xs, _ := linspace(0, 1, chunk_len)
	diff := end_val - start_val
	for i, x := range xs {
		xs[i] = diff*math.Pow(x, slope)/(math.Pow(x, slope)+math.Pow(1-x, slope)) + start_val
	}
	return xs
}

// Return evenly spaced numbers over a specified interval
func linspace(start, stop float64, num int) (ls []float64, err error) {
	if start >= stop {
		return ls, fmt.Errorf("linspace start must not be greater than stop: %v, %v", start, stop)
	}
	if num <= 0 {
		return ls, fmt.Errorf("num must be greater than 0: %v, %v", start, stop)
	}
	ls = make([]float64, num)
	delta := stop / float64(num)
	for i, x := 0, start; x < stop; i, x = i+1, x+delta {
		ls[i] = x
	}
	return ls, nil
}

func minMax(a, b int) (min, max int) {
	if a > b {
		return b, a
	}
	return a, b
}

// GeneratePalette generates palette with the given number of colors
func GeneratePalette(n int) (p []color.Color) {
	p = make([]color.Color, n)
	for i := range p {
		p[i] = palette.Wikipedia.Colors()[i].Color
	}
	return p
}
