package color

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"ledfx/config"
	log "ledfx/logger"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/mazznoer/colorgrad"
	"github.com/ojrac/opensimplex-go"
	"github.com/ritchie46/GOPHY/img2gif"
)

/*func init() {
	go func() {
		log.Logger.Fatalf("Error listening and serving palette handler: %v", http.ListenAndServe(":8740", nil))
	}()
}*/

func PaletteFromPNG(data []byte, resolution int, angle int) (g *Palette, err error) {
	model, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	pString := bytes.NewBuffer([]byte(fmt.Sprintf("linear-palette(%ddeg,", angle)))
	size := model.Bounds().Dx() + model.Bounds().Dy()

	for curX := 0; curX < model.Bounds().Dx(); curX += resolution {
		for curY := 0; curY < model.Bounds().Dy(); curY += resolution {
			// Get the RGB of the current pixel in the image
			R, G, B, _ := model.At(curX, curY).RGBA()
			pString.WriteString(fmt.Sprintf(" rgb(%d, %d, %d) %d%%,", uint8(R>>8), uint8(G>>8), uint8(B>>8), int(math.Round(float64(curX+curY)/float64(size)*float64(100)))))
		}
	}

	pString.Truncate(pString.Len() - 1)
	pString.WriteByte(')')
	return ParsePalette(pString.String())
}

func (g *Palette) WebServe() (link *url.URL, err error) {
	hasher := sha256.New()
	hasher.Write([]byte(g.rawCSS))

	myIP := config.GetSettings().Host
	path := fmt.Sprintf("/palettes/%x", hasher.Sum(nil))

	if link, err = url.Parse(fmt.Sprintf("http://%s:8740%s", myIP, path)); err != nil {
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

func (g *Palette) Raw(width, height int) ([]byte, error) {
	grad, err := colorgrad.NewGradient().Colors(NormalizeColorList(g.colors)...).Build()
	if err != nil {
		return nil, fmt.Errorf("error building palette: %w", err)
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

func (g *Palette) RawNoise(width, height int, seed int64, evalFactor float64) ([]byte, error) {
	grad, err := colorgrad.NewGradient().Colors(NormalizeColorList(g.colors)...).Build()
	if err != nil {
		return nil, fmt.Errorf("error building palette: %w", err)
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
	gr1, err := PaletteFromPNG(data, 2, 90)
	if err != nil {
		return nil, fmt.Errorf("error generating palette from PNG: %w", err)
	}

	imgs := make([]image.Image, numFrames)
	for i := 0; i < numFrames; i++ {
		log.Logger.WithField("context", "Cover Animator").Infof("Computing frame %d", i)
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

func (g *Palette) RawNoiseWithPalette(width, height int, seed int64, pal []color.Color) (*image.Paletted, error) {
	grad, err := colorgrad.NewGradient().Colors(NormalizeColorList(g.colors)...).Build()
	if err != nil {
		return nil, fmt.Errorf("error building palette: %w", err)
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
