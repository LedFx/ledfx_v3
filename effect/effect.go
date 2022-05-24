package effect

import (
	"fmt"
	"ledfx/color"
)

// PixelGenerator is the interface for effect types.
// All effects generate pixels
type PixelGenerator interface {
	Initialize()
	UpdateConfig(c interface{}) (err error)
	AssembleFrame(colors *color.Pixels)
}

type AudioPixelGenerator interface {
	PixelGenerator
	AudioUpdated()
}

type Effect struct {
	Id     string
	Type   string
	Name   string
	Config EffectConfig
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

// Apply mirrors, blur, filters, etc to effect. Should be applied to fresh effect frames.
func (e *Effect) Postprocess(p *color.Pixels) {
	e.applyBkg(p)
	e.applyFlip(p)
	e.applyMirror(p)
	e.applyBlur(p)
	e.applyBrightness(p)
	e.applyGlobals(p)
}

func (e *Effect) applyGlobals(p *color.Pixels) {
	globalHue(p)
	globalBrightness(p)
	globalSaturation(p)
}

func globalBrightness(p *color.Pixels) {}
func globalSaturation(p *color.Pixels) {}
func globalHue(p *color.Pixels) {
	h := globalConfig.Hue //eg.
	fmt.Println(h)
}

func (e *Effect) applyBlur(p *color.Pixels)   {}
func (e *Effect) applyFlip(p *color.Pixels)   {}
func (e *Effect) applyMirror(p *color.Pixels) {}
func (e *Effect) applyBkg(p *color.Pixels)    {}
func (e *Effect) applyBrightness(p *color.Pixels) {
	b := e.Config.Brightness //eg.
	fmt.Println(b)
}
