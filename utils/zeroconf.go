package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ledfx/config"
	"ledfx/device"
	"ledfx/logger"
	"ledfx/virtual"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grandcat/zeroconf"
)

type wledInfo struct {
	Ver  string `json:"ver"`
	Vid  int    `json:"vid"`
	Leds struct {
		Count  int  `json:"count"`
		Rgbw   bool `json:"rgbw"`
		Wv     bool `json:"wv"`
		Cct    bool `json:"cct"`
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

func resolveWledInfo(ip net.IP, id string) bool {
	url := "http://" + ip.String() + "/json/info"

	spaceClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "spacecount-tutorial")

	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	wledInfo1 := wledInfo{}
	jsonErr := json.Unmarshal(body, &wledInfo1)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	fmt.Print(wledInfo1.Name)

	logger.Logger.Debug("New WLED found: ")
	err = device.AddDeviceToConfig(config.Device{
		// TODO: fill in details
		Config: config.DeviceConfig{
			Name:       wledInfo1.Name,
			PixelCount: wledInfo1.Leds.Count,
			IpAddress:  wledInfo1.IP,
		},
		Type: "wled",
		Id:   id,
	}, "goconfig")
	if err != nil {
		logger.Logger.Warn(err)
	}
	var exists bool
	exists, err = virtual.AddDeviceAsVirtualToConfig(config.Virtual{
		Config: config.VirtualConfig{
			CenterOffset:   0,
			FrequencyMax:   15000,
			FrequencyMin:   20,
			IconName:       "wled",
			Mapping:        "span",
			MaxBrightness:  1,
			Name:           wledInfo1.Name,
			PreviewOnly:    false,
			TransitionMode: "Add",
			TransitionTime: 0.4,
		},
		Effect: config.Effect{
			Config: config.EffectConfig{
				BackgroundColor: "#000000",
				Color:           "#eee000",
			},
			Name: "Single Color",
			Type: "singleColor",
		},
		Segments: [][]interface{}{{id, 0, wledInfo1.Leds.Count - 1, false}},
		IsDevice: id,
		Id:       id,
	}, "goconfig")
	if err != nil {
		logger.Logger.Warn(err)
	}
	return exists
}

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
			exists := resolveWledInfo(entry.AddrIPv4[0], entry.ServiceRecord.Instance)
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
