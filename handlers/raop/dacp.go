package raop

import (
	"context"
	"fmt"
	log "ledfx/logger"
	"net/http"
	"time"

	"github.com/grandcat/zeroconf"
)

// DacpClient used to perform DACP operations
type DacpClient struct {
	dacpID       string
	activeRemote string
	ipAddress    string
	port         int
	httpClient   *http.Client
}

func newDacpClient(ipAddress string, port int, dacpID string, activeRemote string) *DacpClient {
	return &DacpClient{ipAddress: ipAddress, port: port, dacpID: dacpID, activeRemote: activeRemote, httpClient: &http.Client{}}
}

func (d *DacpClient) Play() error {
	return d.executeMethod("play")
}

func (d *DacpClient) Pause() error {
	return d.executeMethod("pause")
}

func (d *DacpClient) PlayPause() error {
	return d.executeMethod("playpause")
}

func (d *DacpClient) Stop() error {
	return d.executeMethod("stop")
}

func (d *DacpClient) Next() error {
	return d.executeMethod("nextitem")
}

func (d *DacpClient) executeMethod(method string) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/ctrl-int/1/%s", d.ipAddress, d.port, method), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Active-Remote", d.activeRemote)
	_, err = d.httpClient.Do(req)
	return err
}

// DiscoverDacpClient will try to find the matching DACP client for stream operations
func DiscoverDacpClient(dacpID string, activeRemote string) *DacpClient {
	serviceType := "_dacp._tcp"
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Logger.WithField("context", "DACP Discovery").Errorln("Failed to initialize resolver:", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	defer cancel()
	err = resolver.Browse(ctx, serviceType, "local", entries)
	if err != nil {
		log.Logger.WithField("context", "DACP Discovery").Errorln("Failed to browse:", err.Error())
	}
	log.Logger.WithField("context", "DACP Discovery").Println("searching for DAACP airplay client")
	instanceName := fmt.Sprintf("iTunes_Ctrl_%s", dacpID)
	var entry *zeroconf.ServiceEntry
	foundEntry := make(chan *zeroconf.ServiceEntry)
	// what we do is spin of a goroutine that will process the entries registered in
	// mDNS for our service.  As soon as we detect there is one with an IP4 address
	// we send it off and cancel to stop the searching.
	// there is an issue, https://github.com/grandcat/zeroconf/issues/27 where we
	// could get an entry back without an IP4 addr, it will come in later as an update
	// so we wait until we find the addr, or timeout
	go func(results <-chan *zeroconf.ServiceEntry, foundEntry chan *zeroconf.ServiceEntry) {
		for e := range results {
			if (len(e.AddrIPv4)) > 0 && e.Instance == instanceName {
				foundEntry <- e
				cancel()
			}
		}
	}(entries, foundEntry)

	select {
	case entry = <-foundEntry:
		log.Logger.WithField("context", "DACP Discovery").Println("Found DAACP airplay client")
	case <-ctx.Done():
		log.Logger.WithField("context", "DACP Discovery").Println("DACP airplay client not found")
	}

	if entry == nil {
		log.Logger.WithField("context", "DACP Discovery").Println("no DACP client found, playback control is disabled")
		return nil
	}
	client := newDacpClient(entry.AddrIPv4[0].String(), entry.Port, dacpID, activeRemote)
	return client
}
