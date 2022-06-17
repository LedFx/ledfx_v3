// slow zeroconf on windows
// all XXX seconds, announcements are done automatically
// to forceTrigger the announcement, restart WLED
// as a timing reference for fast zeroconf on windows run:
// dns-sd -B _wled

package util

import (
	"context"
	"ledfx/logger"

	"github.com/grandcat/zeroconf"
)

var resolver *zeroconf.Resolver
var entries chan *zeroconf.ServiceEntry = make(chan *zeroconf.ServiceEntry)
var ctx context.Context
var cancel context.CancelFunc
var running bool

func init() {
	var err error
	resolver, err = zeroconf.NewResolver(nil)
	ctx, cancel = context.WithCancel(context.Background())
	if err != nil {
		logger.Logger.WithField("context", "WLED Scanner").Fatal(err)
	}
}

func handleEntries(results <-chan *zeroconf.ServiceEntry) {
	for entry := range results {
		logger.Logger.WithField("context", "WLED Scanner").Info(entry)

		// err := device.DetectWled(entry.AddrIPv4[0], entry.ServiceRecord.Instance)
		// if Ws != nil && err != nil {
		// 	SendWs(Ws, "info", "New WLED found: "+entry.ServiceRecord.Instance)
		// 	fmt.Print("\n")
		// } else {
		// 	fmt.Println(", but already exsisting in config...")
		// }
		// logger.Logger.Debug(entry.ServiceRecord.Instance)
		// logger.Logger.Debug(" on ")
		// logger.Logger.Debug(entry.AddrIPv4)
	}
}

func EnableScan() error {
	if running {
		return nil
	}
	ctx, cancel = context.WithCancel(context.Background())
	go handleEntries(entries)
	err := resolver.Browse(ctx, "_wled._tcp", "local", entries)
	if err == nil {
		logger.Logger.WithField("context", "WLED Scanner").Info("Enabled WLED Scanner")
		running = true
	} else {
		logger.Logger.WithField("context", "WLED Scanner").Error("Failed to enable WLED scanner;", err)
		cancel()
	}
	return err
}

func DisableScan() {
	if !running {
		return
	}
	logger.Logger.WithField("context", "WLED Scanner").Info("Disabled WLED Scanner")
	cancel()
	running = false
}
