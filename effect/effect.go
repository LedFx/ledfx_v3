package effect

import (
	"fmt"
	"ledfx/color"
)

/*
PixelGenerator is the interface for effect types: all effects generate pixels.
Effects must be computed in HSL space. The color is abstracted and handled outside of the effect
using the effect's palette. Post processing will handle conversion to RGB but requires HSV space
*/
type PixelGenerator interface {
	Initialize(string, int) error
	UpdateConfig(c interface{}) (err error)
	AssembleFrame(colors color.Pixels)
	GetID() string
}

type AudioPixelGenerator interface {
	PixelGenerator
	AudioUpdated()
}

type Effect struct {
	ID         string
	pixelCount int
	palette    [color.PaletteSize]color.Color // the actual palette object, created from the config string
	bkgColor   color.Color                    // parsed background color
	mirror     color.Pixels                   // scratch array used by mirror function
	Config     EffectConfig
}

type EffectConfig struct {
	Intensity     float64 `mapstructure:"intensity" json:"intensity" description:"Visual intensity eg. speed, reactivity" default:"0.5" validate:"gte=0,lte=1"`
	Brightness    float64 `mapstructure:"brightness" json:"brightness" description:"Brightness modifier applied to this effect" default:"1" validate:"gte=0,lte=1"`
	Palette       string  `mapstructure:"palette" json:"palette" description:"Color scheme" default:"RGB" validate:"palette"`
	Blur          float64 `mapstructure:"blur" json:"blur"  description:"Gaussian blur to smoothly blend colors" default:"0.5" validate:"gte=0,lte=1"`
	Flip          bool    `mapstructure:"flip" json:"flip" description:"Reverse the pixels" default:"false" validate:""`
	Mirror        bool    `mapstructure:"mirror" json:"mirror" description:"Mirror the pixels across the center" default:"false" validate:""`
	BkgBrightness float64 `mapstructure:"bkg_brightness" json:"bkg_brightness" description:"Brightness modifier applied to the background color" default:"0.2" validate:"gte=0,lte=1"`
	BkgColor      string  `mapstructure:"bkg_color" json:"bkg_color" description:"Apply a background color" default:"#000000" validate:"color"`
}

func (e *Effect) GetID() string {
	return e.ID
}

// Apply mirrors, blur, filters, etc to effect. Should be applied to fresh effect frames.
func (e *Effect) Postprocess(p color.Pixels) {
	// HSL processes
	e.applyFlip(p)
	e.applyMirror(p)
	color.DarkenPixels(p, e.Config.Brightness)
	e.applyGlobals(p)
	color.ToRGB(p)
	// RGB processes
	e.applyBkg(p)
	e.applyBlur(p)
}

func (e *Effect) applyGlobals(p color.Pixels) {
	color.HueShiftPixels(p, globalConfig.Hue)
	color.DesaturatePixels(p, globalConfig.Saturation)
	color.DarkenPixels(p, globalConfig.Brightness)
}

func (e *Effect) applyBlur(p color.Pixels) {
	if e.Config.Blur == 0 {
		return
	}
}
func (e *Effect) applyFlip(p color.Pixels) {
	if !e.Config.Mirror {
		return
	}
	// in place slice reversal
	for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
		p[i], p[j] = p[j], p[i]
	}
}

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

func (e *Effect) applyBkg(p color.Pixels) {
}
