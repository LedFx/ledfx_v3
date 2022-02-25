package device

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ledfx/config"
	"ledfx/logger"
	"log"
	"net"
	"net/http"
	"time"
)

type WledInfo struct {
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

func DetectWled(ip net.IP, id string) bool {
	// Resolve additional WLED-info
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
	wledInfo1 := WledInfo{}
	fmt.Println(string(body) + "\n")
	jsonErr := json.Unmarshal(body, &wledInfo1)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	// Resolved

	fmt.Print(wledInfo1.Name)
	logger.Logger.Debug("New WLED found: ")

	// Adding Device
	err = AddDeviceToConfig(config.Device{
		// TODO: fill in details
		Config: config.DeviceConfig{
			Name:       wledInfo1.Name,
			PixelCount: wledInfo1.Leds.Count,
			IpAddress:  wledInfo1.IP,
		},
		Type: "wled",
		Id:   id,
	})
	if err != nil {
		logger.Logger.Warn(err)
	}

	// Adding Virtual
	var exists bool
	exists, err = AddDeviceAsVirtualToConfig(config.Virtual{
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
		Id:       id,
		IsDevice: id,
		Segments: [][]interface{}{{id, 0, wledInfo1.Leds.Count - 1, false}},
	})
	if err != nil {
		logger.Logger.Warn(err)
	}
	return exists
}
