package utils

import (
	_ "embed"
	"fmt"
)

//go:embed assets/logo.txt
var logoTxt []byte

func PrintLogo() error {
	s := string(logoTxt)
	fmt.Print(s)
	fmt.Println()
	return nil
}
