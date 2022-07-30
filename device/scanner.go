// slow zeroconf on windows
// all XXX seconds, announcements are done automatically
// to forceTrigger the announcement, restart WLED
// as a timing reference for fast zeroconf on windows run:
// dns-sd -B _wled

// TODO serial device discovery, Art-Poll page 24 https://www.artisticlicence.com/WebSiteMaster/User%20Guides/art-net.pdf

package device

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ledfx/config"
	"ledfx/event"
	"ledfx/logger"
	"net/http"
	"time"

	"github.com/grandcat/zeroconf"
)

var resolver *zeroconf.Resolver
var entries chan *zeroconf.ServiceEntry = make(chan *zeroconf.ServiceEntry)
var ctx context.Context
var cancel context.CancelFunc
var running bool

type wledInfo struct {
	Ver  string `json:"ver"`
	Vid  int    `json:"vid"`
	Leds struct {
		Count  int  `json:"count"`
		Lc     byte `json:"lc"`
		Pwr    int  `json:"pwr"`
		Fps    int  `json:"fps"`
		Maxpwr int  `json:"maxpwr"`
		Maxseg int  `json:"maxseg"`
	} `json:"leds"`
	Str      bool   `json:"str"`
	Name     string `json:"name"`
	Udpport  int    `json:"udpport"`
	Live     bool   `json:"live"`
	Lm       string `json:"lm"`
	Lip      string `json:"lip"`
	Ws       int    `json:"ws"`
	Fxcount  int    `json:"fxcount"`
	Palcount int    `json:"palcount"`
	Wifi     struct {
		Bssid   string `json:"bssid"`
		Rssi    int    `json:"rssi"`
		Signal  int    `json:"signal"`
		Channel int    `json:"channel"`
	} `json:"wifi"`
	Fs struct {
		U   int `json:"u"`
		T   int `json:"t"`
		Pmt int `json:"pmt"`
	} `json:"fs"`
	Ndc      int    `json:"ndc"`
	Arch     string `json:"arch"`
	Core     string `json:"core"`
	Lwip     int    `json:"lwip"`
	Freeheap int    `json:"freeheap"`
	Uptime   int    `json:"uptime"`
	Opt      int    `json:"opt"`
	Brand    string `json:"brand"`
	Product  string `json:"product"`
	Mac      string `json:"mac"`
	IP       string `json:"ip"`
}

func init() {
	var err error
	resolver, err = zeroconf.NewResolver(nil)
	ctx, cancel = context.WithCancel(context.Background())
	if err != nil {
		logger.Logger.WithField("context", "WLED Scanner").Fatal(err)
	}
	event.Subscribe(event.SettingsUpdate, func(e *event.Event) {
		switch config.GetSettings().NoScan {
		case false:
			EnableScan()
		case true:
			DisableScan()
		}
	})
}

func EnableScan() error {
	if running {
		return nil
	}
	ctx, cancel = context.WithCancel(context.Background())

	// handler for the scan results
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			if err := handleEntry(entry); err != nil {
				logger.Logger.WithField("context", "WLED Scanner").Error(err)
			}
		}
	}(entries)

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

func handleEntry(entry *zeroconf.ServiceEntry) error {
	// Discovered WLED service, now need to get additional info
	logger.Logger.WithField("context", "WLED Scanner").Debugf("Found %s at %s", entry.ServiceRecord.Instance, entry.AddrIPv4[0])
	// make request
	url := fmt.Sprintf("http://%s/json/info", entry.AddrIPv4[0].String())
	client := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	// read result
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	info := wledInfo{}
	if err := json.Unmarshal(body, &info); err != nil {
		return err
	}
	// Try to avoid duplication matching IP to other devices
	for id, d := range deviceInstances {
		switch pusher := d.pixelPusher.(type) {
		case *UDP:
			if pusher.config.IP == info.IP {
				logger.Logger.WithField("context", "WLED Scanner").Debugf("Matches IP of %s - Ignoring.", id)
				return nil
			}
		case *ArtNet:
			if pusher.config.IP == info.IP {
				logger.Logger.WithField("context", "WLED Scanner").Debugf("Matches IP of %s - Ignoring.", id)
				return nil
			}
		case *E131:
			for _, ip := range pusher.config.IPs {
				if ip == info.IP {
					logger.Logger.WithField("context", "WLED Scanner").Debugf("Matches IP of %s - Ignoring.", id)
					return nil
				}
			}
		}
	}

	// choose a suitable UDP protocol
	var protocol Protocol
	if info.Leds.Count <= 490 {
		protocol = DRGB
	} else {
		protocol = DNRGB
	}

	// Create the device
	logger.Logger.WithField("context", "WLED Scanner").Infof("Detected WLED %s", info.Name)
	New("",
		"udp_stream",
		map[string]interface{}{
			"name":        info.Name,
			"pixel_count": info.Leds.Count,
		},
		map[string]interface{}{
			"ip":       info.IP,
			"port":     info.Udpport,
			"protocol": protocol,
			"timeout":  3,
		},
	)

	return nil
}
