package utils

import (
	"context"
	"fmt"
	"ledfx/config"
	"log"
	"time"

	"github.com/grandcat/zeroconf"
)

func ScanZeroconf() error {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Println("Failed to initialize resolver:", err.Error())
		return err
	}

	entries := make(chan *zeroconf.ServiceEntry)

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			fmt.Print("New WLED found: ")
			// TODO: check if exists in config already
			config.AddDevice(config.Device{
				// TODO: fill in details
				Config: config.DeviceConfig{
					// IpAddress: entry.AddrIPv4, // convert to string
				},
				Type: "wled",
			}, "goconfig")
			if Ws != nil {
				SendWs(Ws, "info", "New WLED found: "+entry.ServiceRecord.Instance)
			}
			fmt.Print(entry.ServiceRecord.Instance)
			fmt.Print(" on ")
			fmt.Println(entry.AddrIPv4)
		}
		// fmt.Println("No more entries.")
	}(entries)

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*waitTime))
	ctx, _ := context.WithTimeout(context.Background(), time.Second*time.Duration(120))
	// defer cancel()

	err = resolver.Browse(ctx, "_wled._tcp", "local", entries)
	if err != nil {
		log.Println("Failed to browse:", err.Error())
		return err
	}

	<-ctx.Done()
	// Wait some additional time to see debug messages on go routine shutdown.
	time.Sleep(1 * time.Second)
	return nil
}
