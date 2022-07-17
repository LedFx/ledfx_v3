package effect

import (
	"encoding/json"
	"errors"
	"fmt"
	"ledfx/audio"
	"ledfx/color"
	"ledfx/config"
	"ledfx/event"
	"math"
	"time"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

/*
PixelGenerator is the interface for effect types: all effects generate pixels.
Effects must be computed in HSV space. The color is abstracted and handled outside of the effect
using the effect's palette. Post processing will handle conversion to RGB but requires HSV space
*/
type PixelGenerator interface {
	assembleFrame(base *Effect, colors color.Pixels)
}

type AudioPixelGenerator interface {
	PixelGenerator
	AudioUpdated()
}

type Effect struct {
	ID             string
	Type           string
	pixelCount     int
	pixelScaler    float64          // use this to multiply 0-1 float indexes to real integer indexes
	Config         BaseEffectConfig // base config. try to make the effect using only keys from this
	pixelGenerator PixelGenerator   // the effect implementation which produces raw frames
	startTime      time.Time        // time the effect started
	prevFrameTime  time.Time        // time of the previous frams
	deltaStart     time.Duration    // time since effect started
	deltaPrevFrame time.Duration    // time delta since prev frame, so effects can run at constant speed
	palette        *color.Palette   // color palette for the effect. derive single colors from the palette
	blurrer        *color.Blurrer   // blurs the effect
	prevFrame      color.Pixels     // the previous frame, in hsv
	bkgColor       color.Color      // parsed background color
	mirror         color.Pixels     // scratch array used by mirror function
	Ready          bool
}

type BaseEffectConfig struct {
	Intensity     float64 `mapstructure:"intensity" json:"intensity" description:"Visual intensity eg. speed, reactivity" default:"0.5" validate:"gte=0,lte=1"`
	Brightness    float64 `mapstructure:"brightness" json:"brightness" description:"Brightness modifier applied to this effect" default:"1" validate:"gte=0,lte=1"`
	Saturation    float64 `mapstructure:"saturation" json:"saturation" description:"Saturation modifier applied to this effect" default:"1" validate:"gte=0,lte=1"`
	Palette       string  `mapstructure:"palette" json:"palette" description:"Color scheme" default:"RGB" validate:"palette"`
	Blur          float64 `mapstructure:"blur" json:"blur"  description:"Gaussian blur to smoothly blend colors" default:"0.5" validate:"gte=0,lte=1"`
	Flip          bool    `mapstructure:"flip" json:"flip" description:"Reverse the pixels" default:"false" validate:""`
	Mirror        bool    `mapstructure:"mirror" json:"mirror" description:"Mirror the pixels across the center" default:"false" validate:""`
	Decay         float64 `mapstructure:"decay" json:"decay" description:"Apply temporal filtering" default:"0.5" validate:"gte=0,lte=1"`
	HueShift      float64 `mapstructure:"hue_shift" json:"hue_shift" description:"Cycle the colors through time" default:"0" validate:"gte=0,lte=1"`
	BkgBrightness float64 `mapstructure:"bkg_brightness" json:"bkg_brightness" description:"Brightness modifier applied to the background color" default:"0.2" validate:"gte=0,lte=1"`
	BkgColor      string  `mapstructure:"bkg_color" json:"bkg_color" description:"Apply a background color" default:"#000000" validate:"color"`
	FreqMin       int     `mapstructure:"freq_min" json:"freq_min" description:"Lowest audio frequency to react to" default:"20" validate:"gte=20,lte=20000"`
	FreqMax       int     `mapstructure:"freq_max" json:"freq_max" description:"Highest audio frequency to react to" default:"20000" validate:"gte=20,lte=20000"`
}

func (e *Effect) GetID() string {
	return e.ID
}

func (e *Effect) initialize(id string, pixelCount int) {
	e.Ready = false
	e.ID = id
	e.startTime = time.Now()
	e.pixelCount = pixelCount
	e.prevFrame = make(color.Pixels, pixelCount)
	e.mirror = make(color.Pixels, pixelCount)
	e.pixelScaler = float64(pixelCount - 1)
	e.palette = nil
	e.blurrer = nil
	e.Config = globalConfig
}

func (e *Effect) UpdatePixelCount(pixelCount int) error {
	e.initialize(e.ID, pixelCount)
	return e.UpdateBaseConfig(e.Config)
}

/*
Updates the base config of the effect. Config can be given
as EnergyConfig, map[string]interface{}, or raw json.
You can also use a nil to set config to defaults
*/
func (e *Effect) UpdateBaseConfig(c interface{}) (err error) {
	e.Ready = false
	defer func() { e.Ready = true }()
	newConfig := e.Config
	switch t := c.(type) {
	case BaseEffectConfig: // No conversion necessary
		newConfig = c.(BaseEffectConfig)
	case map[string]interface{}: // Decode a map structure
		err = mapstructure.Decode(t, &newConfig)
	case []byte: // Unmarshal a json byte slice
		err = json.Unmarshal(t, &newConfig)
	case nil:
		err = defaults.Set(&newConfig)
	default:
		err = fmt.Errorf("invalid config type: %T %s", t, t)
	}
	if err != nil {
		return err
	}

	// validate all values
	err = validate.Struct(&newConfig)
	if errs, ok := validate.Struct(&newConfig).(validator.ValidationErrors); ok {
		if errs != nil {
			errString := "Validation Errors: "
			for _, err := range errs {
				errString += fmt.Sprintf("Field %s with value %v; ", err.Field(), err.Value())
			}
			return errors.New(errString)
		}
	}

	// create stored properties from new config
	e.updateStoredProperties(newConfig)

	// apply config to effect
	e.Config = newConfig

	// save to config store
	mapConfig := map[string]interface{}{}
	err = mapstructure.Decode(newConfig, &mapConfig)
	if err != nil {
		return err
	}
	err = config.AddEntry(
		e.ID,
		config.EffectEntry{
			ID:         e.ID,
			Type:       e.Type,
			BaseConfig: mapConfig,
		},
	)

	// invoke event
	event.Invoke(event.EffectUpdate,
		map[string]interface{}{
			"id":          e.ID,
			"base_config": mapConfig,
		})
	return err
}

// updates properties and objects which are generated from the config
// eg. melbanks, made using the config frequency range; palette, which is generated from the palette string
func (e *Effect) updateStoredProperties(newConfig BaseEffectConfig) {
	// COLOR AND PALETTE
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
	// MELBANK
	// make sure min and max are ordered properly. doesn't matter in config, but does for audio processing.
	if newConfig.FreqMin > newConfig.FreqMax {
		newConfig.FreqMin, newConfig.FreqMax = newConfig.FreqMax, newConfig.FreqMin
	}
	// we need to make sure that there is a minimum freq difference between min and max.
	// cant be bothered to write a multi field validator, so i'll just do it silently here
	if newConfig.FreqMax-newConfig.FreqMin < 50 {
		// our mel max is limited in config to 20000 but can really go as high as 22050.
		// we'll just add 50 to the max since there's always room there
		newConfig.FreqMax += 50
	}
	// need to register a new melbank if our freqs or audio stream has changed
	if e.Config.FreqMin != newConfig.FreqMin || e.Config.FreqMax != newConfig.FreqMax || e.Config.Intensity != newConfig.Intensity {
		audio.Analyzer.DeleteMelbank(e.ID)
		audio.Analyzer.NewMelbank(e.ID, uint(newConfig.FreqMin), uint(newConfig.FreqMax), newConfig.Intensity)
	}
	// need to register a melbank if the effect doesn't have one yet
	if _, err := audio.Analyzer.GetMelbank(e.ID); err != nil {
		audio.Analyzer.NewMelbank(e.ID, uint(newConfig.FreqMin), uint(newConfig.FreqMax), newConfig.Intensity)
	}
}

// Render a new frame of pixels. Give the previous frame as argument.
// This handles assembling a new frame, then applying mirrors, blur, filters, etc
func (e *Effect) Render(p color.Pixels) {
	if !e.Ready || len(p) != e.pixelCount {
		return
	}
	// These timing variables ensure that temporal effects and filters run at constant
	// speed, irrespective of the effect framerate.
	now := time.Now()
	e.deltaPrevFrame = now.Sub(e.prevFrameTime)
	e.deltaStart = now.Sub(e.startTime)
	e.prevFrameTime = now

	// Overwrite the incoming frame (RGB) with the last frame (HSV) while applying temporal decay
	// for formula explanation, see: https://www.desmos.com/calculator/5qk6xql8bn
	decay := math.Pow(-math.Log((e.Config.Decay*math.E-e.Config.Decay+1)/math.E), 10*e.deltaPrevFrame.Seconds())
	for i := 0; i < e.pixelCount; i++ {
		e.prevFrame[i][2] *= decay
		p[i] = e.prevFrame[i]
	}
	// Assemble new pixels onto the frame
	e.pixelGenerator.assembleFrame(e, p)

	// save the frame to prevFrame
	for i := 0; i < e.pixelCount; i++ {
		e.prevFrame[i] = p[i]
	}

	// HSV processes
	e.applyFlip(p)
	e.applyMirror(p)
	color.HueShiftPixels(p, e.Config.HueShift*e.deltaStart.Seconds())

	// convert p from HSV to RGB using the palette
	for i := 0; i < e.pixelCount; i++ {
		s := p[i][1]
		v := p[i][2]
		p[i] = e.palette.Get(p[i][0])
		p[i] = color.Saturation(p[i], s)
		p[i] = color.Value(p[i], v)
	}

	// RGB processes
	e.applyBkg(p)
	for i := range p {
		p[i] = color.Saturation(p[i], e.Config.Saturation)
		p[i] = color.Value(p[i], e.Config.Brightness)
	}
	e.applyBlur(p)
	e.clamp(p)
	event.Invoke(
		event.EffectRender,
		map[string]interface{}{
			"pixels": p,
		},
	)
}

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
		e.mirror[j] = p[i]
	}

	// assign remaining indices in forward direction
	// eg [0,_,2,_,4,_,6,_,8] -> [0,2,4,6,8,_,_,_,_]
	for i, j := len(p)%2, 0; i < len(p); i, j = i+2, j+1 {
		e.mirror[j] = p[i]
	}

	// assign temp array values to output p
	for i := range p {
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
