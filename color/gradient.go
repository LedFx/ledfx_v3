// Package color provides basic tools for interpreting colors for LedFX
package color

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/mazznoer/colorgrad"
	"github.com/muesli/gamut/palette"
	"github.com/ojrac/opensimplex-go"
	"github.com/ritchie46/GOPHY/img2gif"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	log "ledfx/logger"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"tailscale.com/net/interfaces"

	// Side effects
	_ "golang.org/x/image/riff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/vector"
	_ "golang.org/x/image/vp8"
	_ "golang.org/x/image/vp8l"
	_ "golang.org/x/image/webp"
	_ "image/draw"
	_ "image/jpeg"
)

func init() {
	go func() {
		log.Logger.Fatalf("Error listening and serving gradient handler: %v", http.ListenAndServe(":8740", nil))
	}()
}

var errInvalidGradient = errors.New("invalid gradient")

type Gradient struct {
	mode      string
	angle     int64
	colors    []Color
	positions []float64
	rawCSS    string
}

func (g *Gradient) String() string {
	return g.rawCSS
}

func (g *Gradient) WebServe() (link *url.URL, err error) {
	hasher := sha256.New()
	hasher.Write([]byte(g.rawCSS))

	_, myIP, ok := interfaces.LikelyHomeRouterIP()
	if !ok {
		return nil, errors.New("could not get default outbound IP address")
	}

	path := fmt.Sprintf("/gradients/%x", hasher.Sum(nil))

	if link, err = url.Parse(fmt.Sprintf("http://%s:8740%s", myIP.String(), path)); err != nil {
		return nil, err
	}

	body := gradientBodyBuilder(g.rawCSS)
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		w.Header().Set("content-length", strconv.Itoa(len(body)))
		_, _ = w.Write(body)
	})

	return
}

func (g *Gradient) Raw(width, height int) ([]byte, error) {
	grad, err := colorgrad.NewGradient().Colors(NormalizeColorList(g.colors)...).Build()
	if err != nil {
		return nil, fmt.Errorf("error building gradient: %w", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	fw := float64(width) // We don't want to do a ton of type conversions in the loop
	for x := 0; x < width; x++ {
		col := grad.At(float64(x) / fw)
		for y := 0; y < height; y++ {
			img.Set(x, y, col)
		}
	}
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *Gradient) RawNoise(width, height int, seed int64, evalFactor float64) ([]byte, error) {
	grad, err := colorgrad.NewGradient().Colors(NormalizeColorList(g.colors)...).Build()
	if err != nil {
		return nil, fmt.Errorf("error building gradient: %w", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	noise := opensimplex.NewNormalized(seed)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, grad.At(noise.Eval2(float64(x)*evalFactor, float64(y)*evalFactor)))
		}
	}
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func AnimateAlbumArt(data []byte, width, height, numFrames int) ([]byte, error) {
	gr1, err := GradientFromPNG(data, 2, 90)
	if err != nil {
		return nil, fmt.Errorf("error generating gradient from PNG: %w", err)
	}

	imgs := make([]image.Image, numFrames)
	for i := 0; i < numFrames; i++ {
		log.Logger.WithField("category", "Cover Animator").Infof("Computing frame %d", i)
		noisy, err := gr1.RawNoise(width, height, 996, 0.02)
		if err != nil {
			return nil, fmt.Errorf("error generating raw noisy PNG: %w", err)
		}
		if imgs[i], _, err = image.Decode(bytes.NewReader(noisy)); err != nil {
			return nil, fmt.Errorf("error decoding noisy PNG: %w", err)
		}
	}
	imgsP := img2gif.EncodeImgPaletted(&imgs)

	g := &gif.GIF{
		Image:     make([]*image.Paletted, 0),
		Delay:     make([]int, 0),
		LoopCount: 0,
		Config: image.Config{
			Width:  width,
			Height: height,
		},
	}
	for _, i := range imgsP {
		g.Image = append(g.Image, i)
		g.Delay = append(g.Delay, 0)
	}

	buf := bytes.NewBuffer(make([]byte, 0))

	if err := gif.EncodeAll(buf, g); err != nil {
		return nil, fmt.Errorf("error encoding GIF: %w", err)
	}

	return buf.Bytes(), nil
}

func (g *Gradient) RawNoiseWithPalette(width, height int, seed int64, pal []color.Color) (*image.Paletted, error) {
	grad, err := colorgrad.NewGradient().Colors(NormalizeColorList(g.colors)...).Build()
	if err != nil {
		return nil, fmt.Errorf("error building gradient: %w", err)
	}

	img := image.NewPaletted(image.Rect(0, 0, width, height), pal)
	noise := opensimplex.NewNormalized(seed)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, grad.At(noise.Eval2(float64(x)*0.02, float64(y)*0.02)))
		}
	}
	return img, nil
}

func NewGradient(gs string) (g *Gradient, err error) {
	predef, isPredef := LedFxGradients[gs]
	if isPredef {
		return parseGradient(predef)
	}
	return parseGradient(gs)
}

func GradientFromPNG(data []byte, resolution int, angle int) (g *Gradient, err error) {
	model, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	gString := bytes.NewBuffer([]byte(fmt.Sprintf("linear-gradient(%ddeg,", angle)))
	size := model.Bounds().Dx() + model.Bounds().Dy()

	for curX := 0; curX < model.Bounds().Dx(); curX += resolution {
		for curY := 0; curY < model.Bounds().Dy(); curY += resolution {
			// Get the RGB of the current pixel in the image
			R, G, B, _ := model.At(curX, curY).RGBA()
			gString.WriteString(fmt.Sprintf(" rgb(%d, %d, %d) %d%%,", uint8(R>>8), uint8(G>>8), uint8(B>>8), int(math.Round(float64(curX+curY)/float64(size)*float64(100)))))
		}
	}

	gString.Truncate(gString.Len() - 1)
	gString.WriteByte(')')
	return parseGradient(gString.String())
}

/*
Parses gradient from string of format eg.
"linear-gradient(90deg, rgb(100, 0, 255) 0%, #800000 50%, #ec77ab 100%)"
where each color is associated with a % value for its position in the gradient
each color can be hex or rgb format
*/
func parseGradient(gs string) (g *Gradient, err error) {
	g = &Gradient{
		rawCSS: gs,
	}

	var splits []string
	gs = strings.ToLower(gs)
	gs = strings.ReplaceAll(gs, " ", "")
	splits = strings.SplitN(gs, "(", 2)
	mode := splits[0]
	g.mode = strings.TrimSuffix(mode, "-gradient")
	angleColorPos := splits[1]
	angleColorPos = strings.TrimRight(angleColorPos, ")")
	splits = strings.SplitN(angleColorPos, ",", 2)
	if (len(splits) != 2) || !strings.HasSuffix(splits[0], "deg") {
		return nil, errInvalidGradient
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
			err = errInvalidGradient
		}
		if err != nil {
			break
		}
		g.colors[i] = c
		g.positions[i] = p
	}
	if err != nil {
		err = errInvalidGradient
		return nil, err
	}
	return g, err
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
