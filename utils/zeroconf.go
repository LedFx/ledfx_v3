package utils

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/grandcat/zeroconf"
)

var (
	service  = flag.String("service", "_wled._tcp", "Set the service category to look for devices.")
	domain   = flag.String("domain", "local", "Set the search domain. For local networks, default is fine.")
	waitTime = flag.Int("wait", 120, "Duration in [s] to run discovery.")
)

func ScanZeroconf() error {
	// fmt.Println("Scanning... ")
	flag.Parse()

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Println("Failed to initialize resolver:", err.Error())
		return err
	}

	entries := make(chan *zeroconf.ServiceEntry)

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			fmt.Print("New WLED found: ")
			SendWs(Ws, "info", "New WLED found: "+entry.ServiceRecord.Instance)
			fmt.Print(entry.ServiceRecord.Instance)
			fmt.Print(" on ")
			fmt.Println(entry.AddrIPv4)
		}
		// fmt.Println("No more entries.")
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*waitTime))
	defer cancel()
	err = resolver.Browse(ctx, *service, *domain, entries)
	if err != nil {
		log.Println("Failed to browse:", err.Error())
		return err
	}

	<-ctx.Done()
	// Wait some additional time to see debug messages on go routine shutdown.
	time.Sleep(1 * time.Second)
	return nil
}
