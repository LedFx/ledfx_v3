package effect

import (
	"fmt"
	"ledfx/color"
)

// PixelGenerator is the interface for effect types.
// All effects generate pixels
type PixelGenerator interface {
	Initialize(id string)
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
	Intensity     float64       `mapstructure:"intensity" json:"intensity"`
	Brightness    float64       `mapstructure:"brightness" json:"brightness"`
	Palette       color.Palette `mapstructure:"palette" json:"palette"`
	Blur          float64       `mapstructure:"blur" json:"blur"`
	Flip          bool          `mapstructure:"flip" json:"flip"`
	Mirror        bool          `mapstructure:"mirror" json:"mirror"`
	BkgBrightness float64       `mapstructure:"bkg_brightness" json:"bkg_brightness"`
	BkgColor      string        `mapstructure:"bkg_color" json:"bkg_color"`
	// GradientName  string  `mapstructure:"gradient_name" json:"gradient_name"`
	// Color         string  `mapstructure:"color" json:"color"`
}

// Points to a virtual, where this effect will send its pixels to
type EffectOutput struct {
	Id     string `mapstructure:"id" json:"id"`         // Virtual ID
	Active string `mapstructure:"active" json:"active"` // Is this output active
}

// Apply mirrors, blur, filters, etc to effect. Should be applied to fresh effect frames.
func (e *Effect) Postprocess(p *color.Pixels) {
	e.applyBkg(p)
	e.applyMirror(p)
	e.applyFlip(p)
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
