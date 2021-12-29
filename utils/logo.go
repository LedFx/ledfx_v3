package utils

import (
	_ "embed"
	"fmt"

	"github.com/nathan-fiscaletti/consolesize-go"
)

//go:embed assets/logo.txt
var logoTxt []byte

//go:embed assets/logo-sm.txt
var smLogoTxt []byte

func PrintLogo() error {
	cols, _ := consolesize.GetConsoleSize()
	var s string
	if cols >= 125 {
		s = string(logoTxt)
	} else if cols >= 52 {
		s = string(smLogoTxt)
	} else {
		return nil
	}
	fmt.Print(s)
	fmt.Println()
	return nil
}
