package constants

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/qeesung/image2ascii/convert"
)

var LOGO_FILENAME = "logo.png"

func GetLogoPath() (file string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	file = filepath.Join(wd, "static/"+LOGO_FILENAME)
	return
}

func PrintLogo() error {

	// Create convert options
	convertOptions := convert.DefaultOptions
	convertOptions.FixedWidth = 100
	convertOptions.FixedHeight = 25

	// Create the image converter
	converter := convert.NewImageConverter()
	logoPath, err := GetLogoPath()
	if err != nil {
		return err
	}
	fmt.Print(converter.ImageFile2ASCIIString(logoPath, &convertOptions))
	return nil
}
