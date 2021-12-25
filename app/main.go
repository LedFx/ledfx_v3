package main

import (
	"ledfx/ledfx/cli"
	"log"
)

func main() {
	err := cli.InitCli()

	if err != nil {
		log.Fatal(err)
	}
}
