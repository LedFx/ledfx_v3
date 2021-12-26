package constants

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var LOGO_FILENAME = "logo.txt"

func getLogoPath() (file string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	file = filepath.Join(wd, "static/"+LOGO_FILENAME)
	return
}

func PrintLogo() error {
	path, err := getLogoPath()
	if err != nil {
		return err
	}
	log.Println(path)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
	}()

	b, err := ioutil.ReadAll(file)
	if err == nil {
		s := string(b)
		fmt.Print(s)
	}
	return nil
}
