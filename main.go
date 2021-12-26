package main

import (
	"fmt"
	"ledfx/api"
	"ledfx/config"
	"ledfx/constants"
)

func init() {
	config.InitFlags()
}

func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		return
	}

	if conf.Version {
		fmt.Println("LedFx " + constants.VERSION)
		return
	}

	// TODO: handle other flags
	/**
	  OpenUi
	  Verbose
	  VeryVerbose
	  Host
	  Offline
	  SentryCrash
	*/

	err = api.InitApi(conf.Port)
	if err != nil {
		return
	}

}
