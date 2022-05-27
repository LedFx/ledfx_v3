package virtual

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
	PixelCount int `mapstructure:"pixel_count" json:"pixel_count" description:"Number of pixels" validate:"required,gte=0"`
	// length     float64 `mapstructure:"length" json:"length" description:"Length of strip in meters. Hint: Pixels ÷ PixelDensity" validate:"required,gte=0.01"`
}

type TwoDimensionerConfig struct {
	RowCount       int    `mapstructure:"row_count" json:"row_count" description:"Number of rows (horizontal direction)" validate:"required,gte=1"`
	ColCount       int    `mapstructure:"col_count" json:"col_count" description:"Number of columns (vertical direction)" validate:"required,gte=1"`
	StartCorner    string `mapstructure:"start_corner" json:"start_corner" description:"Corner the LEDs start in" default:"top_left" validate:"oneof=top_left top_right bottom_left bottom_right"`
	StartDirection string `mapstructure:"start_direction" json:"start_direction" description:"Direction the LEDs first go in" default:"horizontal" validate:"oneof=horizontal vertical"`
	Arrangement    string `mapstructure:"arrangement" json:"arrangement" description:"How the LEDs are linked between rows/cols" default:"snake" validate:"oneof=snake zigzag"`
	Mapping        string `mapstructure:"mapping" json:"mapping" description:"How 1D effects are mapped onto the matrix" default:"circle" validate:"oneof=circle spiral 1D_multiply 1D_horizontal 1D_vertical none"`
}

func (d *ZeroDimensioner) NumPixels() int { return 1 }

func (d *OneDimensioner) NumPixels() int { return d.Config.PixelCount }

func (d *TwoDimensioner) NumPixels() int {
	switch d.Config.Mapping {
	case "1Dmultiply":
		if d.Config.RowCount < d.Config.ColCount {
			return d.Config.RowCount
		} else {
			return d.Config.ColCount
		}
	case "1D_horizontal":
		return d.Config.RowCount
	case "1D_vertical":
		return d.Config.ColCount
	default:
		return d.Config.RowCount * d.Config.ColCount
	}
}

func (d *ZeroDimensioner) Transform(p color.Pixels) color.Pixels { return p }

func (d *OneDimensioner) Transform(p color.Pixels) color.Pixels { return p }

func (d *TwoDimensioner) Transform(p color.Pixels) color.Pixels {
	switch d.Config.Mapping {
	case "1Dmultiply":
		if d.Config.RowCount < d.Config.ColCount {
			return p // TODO
		} else {
			return p
		}
	case "1D_horizontal":
		return p
	case "1D_vertical":
		return p
	default:
		return p
	}
}
