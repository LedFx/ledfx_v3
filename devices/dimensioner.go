package device

import "ledfx/color"

// Defines the number of pixels required for a dimension
// Applies a transformation to map the pixels to that dimension
type PixelDimensioner interface {
	NumPixels() int
	Transform(p color.Pixels) color.Pixels
}

type ZeroDimensioner struct {
	Config ZeroDimensionerConfig
}

type OneDimensioner struct {
	Config OneDimensionerConfig
}

type TwoDimensioner struct {
	Config TwoDimensionerConfig
}

type ZeroDimensionerConfig struct{}

type OneDimensionerConfig struct {
	pixelCount int `mapstructure:"pixel_count" json:"pixel_count" description:"Number of pixels" validate:"required,gte=0"`
	// length     float64 `mapstructure:"length" json:"length" description:"Length of strip in meters. Hint: Pixels ÷ PixelDensity" validate:"required,gte=0.01"`
}

type TwoDimensionerConfig struct {
	rowCount       int    `mapstructure:"row_count" json:"row_count" description:"Number of rows (horizontal direction)" validate:"required,gte=1"`
	colCount       int    `mapstructure:"col_count" json:"col_count" description:"Number of columns (vertical direction)" validate:"required,gte=1"`
	startCorner    string `mapstructure:"start_corner" json:"start_corner" description:"Corner the LEDs start in" default:"top_left" validate:"oneof=top_left top_right bottom_left bottom_right"`
	startDirection string `mapstructure:"start_direction" json:"start_direction" description:"Direction the LEDs first go in" default:"horizontal" validate:"oneof=horizontal vertical"`
	arrangement    string `mapstructure:"arrangement" json:"arrangement" description:"How the LEDs are linked between rows/cols" default:"snake" validate:"oneof=snake zigzag"`
	mapping        string `mapstructure:"mapping" json:"mapping" description:"How 1D effects are mapped onto the matrix" default:"circle" validate:"oneof=circle spiral 1D_multiply 1D_horizontal 1D_vertical none"`
}

func (d *ZeroDimensioner) NumPixels() int { return 1 }

func (d *OneDimensioner) NumPixels() int { return d.Config.pixelCount }

func (d *TwoDimensioner) NumPixels() int {
	switch d.Config.mapping {
	case "1Dmultiply":
		if d.Config.rowCount < d.Config.colCount {
			return d.Config.rowCount
		} else {
			return d.Config.colCount
		}
	case "1D_horizontal":
		return d.Config.rowCount
	case "1D_vertical":
		return d.Config.colCount
	default:
		return d.Config.rowCount * d.Config.colCount
	}
}

func (d *ZeroDimensioner) Transform(p color.Pixels) color.Pixels { return p }

func (d *OneDimensioner) Transform(p color.Pixels) color.Pixels { return p }

func (d *TwoDimensioner) Transform(p color.Pixels) color.Pixels {
	switch d.Config.mapping {
	case "1Dmultiply":
		if d.Config.rowCount < d.Config.colCount {
			return d.Config.rowCount
		} else {
			return d.Config.colCount
		}
	case "1D_horizontal":
		return d.Config.rowCount
	case "1D_vertical":
		return d.Config.colCount
	default:
		return d.Config.rowCount * d.Config.colCount
	}
}
