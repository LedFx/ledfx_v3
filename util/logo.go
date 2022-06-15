//go:generate goversioninfo -icon=assets/logo.ico
package util

import (
	_ "embed"
	"fmt"
)

//go:embed assets/logo.ico
var logo []byte

//go:embed assets/logo.txt
var logoTxt []byte

//go:embed assets/logo-sm.txt
var smLogoTxt []byte

func PrintLogo() {
	fmt.Println()
	// fmt.Print(string(logoTxt))
	fmt.Print(string(smLogoTxt))
	fmt.Println()
}
