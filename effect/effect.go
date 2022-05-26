package effect

import (
	"encoding/json"
	"fmt"
	"ledfx/color"
	"math"
	"time"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
)

/*
PixelGenerator is the interface for effect types: all effects generate pixels.
Effects must be computed in HSL space. The color is abstracted and handled outside of the effect
using the effect's palette. Post processing will handle conversion to RGB but requires HSL space
*/
type PixelGenerator interface {
	Initialize(id string, pixelCount int) error
	UpdateConfig(c interface{}) (err error)
	UpdateExtraConfig(c interface{}) (err error)
	Render(color.Pixels)
	assembleFrame(colors color.Pixels)
	GetID() string
}

type AudioPixelGenerator interface {
	PixelGenerator
	AudioUpdated()
}

type Effect struct {
	ID            string
	pixelCount    int
	startTime     time.Time
	prevFrameTime time.Time
	palette       *color.Palette
	blurrer       *color.Blurrer
	prevFrame     color.Pixels
	bkgColor      color.Color  // parsed background color
	mirror        color.Pixels // scratch array used by mirror function
	Config        BaseEffectConfig
}

type BaseEffectConfig struct {
	Intensity     float64 `mapstructure:"intensity" json:"intensity" description:"Visual intensity eg. speed, reactivity" default:"0.5" validate:"gte=0,lte=1"`
	Brightness    float64 `mapstructure:"brightness" json:"brightness" description:"Brightness modifier applied to this effect" default:"1" validate:"gte=0,lte=1"`
	Saturation    float64 `mapstructure:"saturation" json:"saturation" description:"Saturation modifier applied to this effect" default:"1" validate:"gte=0,lte=1"`
	Palette       string  `mapstructure:"palette" json:"palette" description:"Color scheme" default:"RGB" validate:"palette"`
	Blur          float64 `mapstructure:"blur" json:"blur"  description:"Gaussian blur to smoothly blend colors" default:"0.5" validate:"gte=0,lte=1"`
	Flip          bool    `mapstructure:"flip" json:"flip" description:"Reverse the pixels" default:"false" validate:""`
	Mirror        bool    `mapstructure:"mirror" json:"mirror" description:"Mirror the pixels across the center" default:"false" validate:""`
	Decay         float64 `mapstructure:"decay" json:"decay" description:"Apply temporal filtering" default:"0" validate:"gte=0,lte=1"`
	HueShift      float64 `mapstructure:"hue_shift" json:"hue_shift" description:"Cycle the colors through time" default:"1" validate:"gte=0,lte=1"`
	BkgBrightness float64 `mapstructure:"bkg_brightness" json:"bkg_brightness" description:"Brightness modifier applied to the background color" default:"0.2" validate:"gte=0,lte=1"`
	BkgColor      string  `mapstructure:"bkg_color" json:"bkg_color" description:"Apply a background color" default:"#000000" validate:"color"`
}

func (e *Effect) GetID() string {
	return e.ID
}

func (e *Effect) Initialize(id string, pixelCount int) error {
	e.ID = id
	e.startTime = time.Now()
	e.pixelCount = pixelCount
	e.prevFrame = make(color.Pixels, pixelCount)
	e.mirror = make(color.Pixels, pixelCount)
	e.palette = nil
	e.blurrer = nil
	err := defaults.Set(&e.Config)
	if err != nil {
		return err
	}
	return e.UpdateConfig(e.Config)
}

/*
Updates the config of the effect. Config can be given
as EnergyConfig, map[string]interface{}, or raw json
*/
func (e *Effect) UpdateConfig(c interface{}) (err error) {
	newConfig := e.Config
	switch t := c.(type) {
	case BaseEffectConfig: // No conversion necessary
		newConfig = c.(BaseEffectConfig)
	case map[string]interface{}: // Decode a map structure
		err = mapstructure.Decode(t, &newConfig)
	case []byte: // Unmarshal a json byte slice
		err = json.Unmarshal(t, &newConfig)
	default:
		err = fmt.Errorf("invalid config type: %s", t)
	}
	if err != nil {
		return err
	}

	// validate all values
	err = validate.Struct(&newConfig)
	if err != nil {
		return err
	}

	// update any stored properties that are based on the config
	// creating a new palette is expensive, should only be done if changed
	if e.palette == nil || e.Config.Palette != newConfig.Palette {
		e.palette, _ = color.NewPalette(newConfig.Palette)
	}
	// parsing a color is cheap, just do it every time
	e.bkgColor, _ = color.NewColor(e.Config.BkgColor)
	// blur needs new blurrer if changed
	if e.blurrer == nil || e.Config.Blur != newConfig.Blur {
		e.blurrer = color.NewBlurrer(e.pixelCount, newConfig.Blur)
	}

	// apply config to effect
	e.Config = newConfig
	return nil
}

// Effect implementation must override this method
func (e *Effect) assembleFrame(p color.Pixels) {}

// Effect implementation may override this method
func (e *Effect) UpdateExtraConfig(c interface{}) (err error) { return nil }

// Render a new frame of pixels. Give the previous frame as argument.
// This handles assembling a new frame, then applying mirrors, blur, filters, etc
func (e *Effect) Render(p color.Pixels) {
	// These timing variables ensure that temporal effects and filters run at constant
	// speed, irrespective of the effect framerate.
	now := time.Now()
	deltaPrevFrame := now.Sub(e.prevFrameTime)
	deltaStart := now.Sub(e.startTime)
	e.prevFrameTime = now

	// Overwrite the incoming frame (RGB) with the last frame (HSL) while applying temporal decay
	// for formula explanation, see: https://www.desmos.com/calculator/5qk6xql8bn
	decay := math.Pow(-math.Log((e.Config.Decay*math.E-e.Config.Decay+1)/math.E), 10*deltaPrevFrame.Seconds())
	for i := 0; i < e.pixelCount; i++ {
		e.prevFrame[i][2] *= decay
		p[i] = e.prevFrame[i]
	}
	// Assemble new pixels onto the frame
	e.assembleFrame(p)

	// HSL processes
	e.applyFlip(p)
	e.applyMirror(p)
	color.HueShiftPixels(p, e.Config.HueShift*deltaStart.Seconds())

	// convert p from HSL to RGB using the palette
	for i := 0; i < e.pixelCount; i++ {
		p[i] = e.palette.Get(p[i][0])
	}

	// RGB processes
	e.applyBkg(p)
	color.DesaturatePixels(p, e.Config.Saturation)
	color.DarkenPixels(p, e.Config.Brightness)
	e.applyBlur(p)
	e.clamp(p)

	// save the frame to prevFrame
	for i := 0; i < e.pixelCount; i++ {
		e.prevFrame[i] = p[i]
	}
}

// TODO pseudo-gaussian blur
func (e *Effect) applyBlur(p color.Pixels) {
	if e.Config.Blur == 0 {
		return
	}
	e.blurrer.BoxBlur(p)
}

// Reverses the pixels
func (e *Effect) applyFlip(p color.Pixels) {
	if !e.Config.Mirror {
		return
	}
	// in place slice reversal
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
}

// Mirrors pixels down the centre
func (e *Effect) applyMirror(p color.Pixels) {
	if !e.Config.Mirror {
		return
	}
	// assign indices from end in reverse direction
	// eg [_,1,_,3,_,5,_,7,_] -> [_,_,_,_,_,7,5,3,1]
	for i, j := len(p)-1, len(p)/2; i >= 0; i, j = i-2, j+1 {
		fmt.Println(j, i)
		e.mirror[j] = p[i]
	}

	// assign remaining indices in forward direction
	// eg [0,_,2,_,4,_,6,_,8] -> [0,2,4,6,8,_,_,_,_]
	for i, j := len(p)%2, 0; i < len(p); i, j = i+2, j+1 {
		e.mirror[j] = p[i]
	}

	// assign temp array values to output p
	for i, _ := range p {
		p[i] = e.mirror[i]
	}
}

// mixes a background colour
func (e *Effect) applyBkg(p color.Pixels) {
	for i := 0; i < e.pixelCount; i++ {
		p[i][0] += e.bkgColor[0] * e.Config.BkgBrightness
		p[i][1] += e.bkgColor[1] * e.Config.BkgBrightness
		p[i][2] += e.bkgColor[2] * e.Config.BkgBrightness
	}
}

// makes sure all pixel values are between 0-1
func (e *Effect) clamp(p color.Pixels) {
	for i := 0; i < e.pixelCount; i++ {
		for k := 0; k < 3; k++ {
			if p[i][k] > 1 {
				p[i][k] = 1
			}
			if p[i][k] < 0 {
				p[i][k] = 0
			}
		}
	}
}
