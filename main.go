package main

import (
	"fmt"
	"ledfx/config"
	"ledfx/constants"
	"ledfx/devices"
	"log"
)

func init() {
	err := config.InitFlags()

	if err != nil {
		log.Fatal(err)
	}
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

	err = constants.PrintLogo()
	if err != nil {
		log.Fatal(err)
	}
	// err = api.InitApi(conf.Port)
	if err != nil {
		log.Fatal(err)
	}

	// REMOVEME: testing only
	devices.SendUdpPacket()

}
