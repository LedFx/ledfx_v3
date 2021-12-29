package utils

import (
	_ "embed"
	"fmt"
)

//go:embed assets/logo.txt
var logoTxt []byte

// TODO: make this responsive to terminal size `$ stty size`
func PrintLogo() error {
	s := string(logoTxt)
	fmt.Print(s)
	fmt.Println()
	return nil
}
