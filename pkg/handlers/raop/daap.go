package raop

import (
	"encoding/binary"
	"fmt"
	pretty "github.com/fatih/color"
)

var (
	lastSong string
)

const (
	goByte   = "byte"
	goString = "string"
)

type contentType struct {
	code  string
	name  string
	cType string
}

// based on: https://github.com/kylewelsby/daap/blob/master/index.js
func parseDaap(daap []byte) map[string]interface{} {
	i := 8
	parsedData := make(map[string]interface{})
	for i < len(daap) {
		itemType := string(daap[i : i+4])
		itemLength := int(binary.BigEndian.Uint32(daap[i+4 : i+8]))
		if itemLength != 0 {
			data := daap[i+8 : i+8+itemLength]
			contentType := getContentType(itemType)
			switch contentType.cType {
			case goByte:
				parsedData[contentType.name] = data[0]
			case goString:
				parsedData[contentType.name] = string(data)
			}
		}
		i = i + itemLength + 8
	}
	logTrackInfo(parsedData)
	return parsedData
}

func logTrackInfo(data map[string]interface{}) {
	curSong := data["dmap.itemname"]
	switch {
	case curSong == nil:
		fallthrough
	case curSong == lastSong:
		fallthrough
	case curSong == "Loading…":
		return
	default:
		if _, ok := curSong.(string); !ok {
			return
		}

		artist, ok := data["daap.songartist"]
		if !ok {
			return
		}

		if _, ok = artist.(string); !ok {
			return
		}

		lastSong = curSong.(string)

		PrettyPrintTrackString(curSong.(string), artist.(string))
	}
}

func PrettyPrintTrackString(song, artist string) {
	_, _ = pretty.Set(pretty.BgHiCyan, pretty.FgBlack, pretty.Bold).Print("🎵 Now playing")
	pretty.Unset()

	_, _ = pretty.Set(pretty.FgHiWhite, pretty.Bold).Print(" ➜ ")
	pretty.Unset()

	_, _ = pretty.Set(pretty.BgMagenta, pretty.FgWhite, pretty.Bold).Printf("%s by %s", song, artist)
	pretty.Unset()
	fmt.Println()
}

// EncodeDaap will take a map and encode it in daap format
func EncodeDaap(dataToEncode map[string]interface{}) ([]byte, error) {
	var buf []byte
	// can't find why, but I found needed to add a padding before the daap data
	padding := make([]byte, 8)
	binary.BigEndian.PutUint64(padding, uint64(0))
	// we first add a padding, from the sample I captured from apple it started
	// with these bytes: 109, 108, 105, 116, 0, 0, 6, 17 but for now I'll just add 0s
	buf = append(buf, padding...)
	for k, v := range dataToEncode {
		ct := getContentTypeForName(k)
		itemType := ct.code
		field := []byte(itemType)
		// format is dataType, dataLength, data
		buf = append(buf, field...)
		var length []byte
		var data []byte
		if ct.cType == goByte {
			data = make([]byte, 1)
			data[0] = v.(uint8)
			length = make([]byte, 4)
			// length of type byte is... one byte :)
			binary.BigEndian.PutUint32(length, 1)
		}
		if ct.cType == goString {
			data = []byte(fmt.Sprintf("%s", v))
			length = make([]byte, 4)
			// add in the length of the data we want to encode
			binary.BigEndian.PutUint32(length, uint32(len(data)))
		}
		buf = append(buf, length...)
		buf = append(buf, data...)
	}

	return buf, nil
}

func getContentType(code string) contentType {
	ct := contentType{}
	// there is a whole TON of types that can come back
	// only parse out the ones we are interested in for now
	switch code {
	case "mikd":
		ct.cType = goByte
		ct.code = "mikd"
		ct.name = "dmap.itemkind"
	case "asal":
		ct.cType = goString
		ct.code = "asal"
		ct.name = "daap.songalbum"
	case "asar":
		ct.cType = goString
		ct.code = "asar"
		ct.name = "daap.songartist"
	case "minm":
		ct.cType = goString
		ct.code = "minm"
		ct.name = "dmap.itemname"
	}
	return ct
}

func getContentTypeForName(name string) contentType {
	ct := contentType{}
	// there is a whole TON of types that can come back
	// only parse out the ones we are interested in for now
	switch name {
	case "dmap.itemkind":
		ct.cType = goByte
		ct.code = "mikd"
		ct.name = "dmap.itemkind"
	case "daap.songalbum":
		ct.cType = goString
		ct.code = "asal"
		ct.name = "daap.songalbum"
	case "daap.songartist":
		ct.cType = goString
		ct.code = "asar"
		ct.name = "daap.songartist"
	case "dmap.itemname":
		ct.cType = goString
		ct.code = "minm"
		ct.name = "dmap.itemname"
	}
	return ct
}
