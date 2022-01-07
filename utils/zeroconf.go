// slow zeroconf on windows
// all XXX seconds, announcements are done automatically
// to forceTrigger the announcement, restart WLED
// as a timing reference for fast zeroconf on windows run:
// dns-sd -B _wled

package utils

import (
	"context"
	"fmt"
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
			fmt.Print("WLED found: ")

			exists := device.DetectWled(entry.AddrIPv4[0], entry.ServiceRecord.Instance)
			if Ws != nil && !exists {
				SendWs(Ws, "info", "New WLED found: "+entry.ServiceRecord.Instance)
				fmt.Print("\n")
			} else {
				fmt.Println(", but already exsisting in config...")
			}
			logger.Logger.Debug(entry.ServiceRecord.Instance)
			logger.Logger.Debug(" on ")
			logger.Logger.Debug(entry.AddrIPv4)
		}
	}(entries)

	ctx := context.Background()

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
