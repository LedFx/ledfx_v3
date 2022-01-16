package codec

import (
	"bytes"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
)

func AdjustAudio(raw []byte, vol float64) []byte {
	if vol == 1 {
		return raw
	}
	adjusted := new(bytes.Buffer)
	for i := 0; i < len(raw); i += 2 {
		var val int16
		b := raw[i : i+2]
		buf := bytes.NewReader(b)
		if err := binary.Read(buf, binary.LittleEndian, &val); err != nil {
			log.Warnf("Error reading binary data: %v\n", err)
			return raw
		}
		val = int16(vol * float64(val))
		val = min(32767, val)
		val = max(-32767, val)
		_ = binary.Write(adjusted, binary.LittleEndian, val)
	}

	return adjusted.Bytes()
}

func min(a, b int16) int16 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int16) int16 {
	if a > b {
		return a
	}
	return b
}
