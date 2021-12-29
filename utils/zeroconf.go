package utils

import (
	"context"
	"fmt"
	"ledfx/config"
	"ledfx/device"
	"ledfx/logger"
	"time"

	"github.com/grandcat/zeroconf"
)

func ScanZeroconf() error {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		logger.Logger.Warn("Failed to initialize resolver:", err.Error())
		return err
	}

	entries := make(chan *zeroconf.ServiceEntry)

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			logger.Logger.Debug("New WLED found: ")
			err = device.AddDeviceToConfig(config.Device{
				// TODO: fill in details
				Config: config.DeviceConfig{
					Name:      entry.ServiceRecord.Instance,
					IpAddress: fmt.Sprintf("%q", entry.AddrIPv4[0]), // convert to string
				},
				Type: "wled",
			}, "goconfig")
			if err != nil {
				logger.Logger.Warn(err)
			}
			if Ws != nil {
				SendWs(Ws, "info", "New WLED found: "+entry.ServiceRecord.Instance)
			}
			logger.Logger.Debug(entry.ServiceRecord.Instance)
			logger.Logger.Debug(" on ")
			logger.Logger.Debug(entry.AddrIPv4)
		}
		// fmt.Println("No more entries.")
	}(entries)

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*waitTime))
	ctx := context.Background()
	// defer cancel()

	err = resolver.Browse(ctx, "_wled._tcp", "local", entries)
	if err != nil {
		logger.Logger.Warn("Failed to browse:", err.Error())
		return err
	}

	<-ctx.Done()
	// Wait some additional time to see debug messages on go routine shutdown.
	time.Sleep(1 * time.Second)
	return nil
}
