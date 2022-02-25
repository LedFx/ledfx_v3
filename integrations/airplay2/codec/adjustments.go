package codec

import "encoding/binary"

func NormalizeAudio(audioBytes []byte, volume float64) {
	if volume == 1 {
		return
	}
	for i := 0; i < len(audioBytes); i += 2 {
		binary.LittleEndian.PutUint16(audioBytes[i:i+2], uint16(max(-32767, min(32767, int16(volume*float64(int16(binary.LittleEndian.Uint16(audioBytes[i:i+2]))))))))
	}
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
