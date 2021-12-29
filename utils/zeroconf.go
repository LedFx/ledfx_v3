package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ledfx/config"
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

func resolveWledInfo(ip net.IP) {
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

	fmt.Println(wledInfo1.Name)

	config.AddDevice(config.Device{
		// TODO: fill in details
		Config: config.DeviceConfig{
			Name:       wledInfo1.Name,
			PixelCount: wledInfo1.Leds.Count,
			// Id: entry.ServiceRecord.Instance,
			IpAddress: wledInfo1.IP, // fmt.Sprintf("%s", entry.AddrIPv4[0]), // convert to string
		},
		Type: "wled",
	}, "goconfig")
}

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
			resolveWledInfo(entry.AddrIPv4[0])

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
	ctx := context.Background()
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
