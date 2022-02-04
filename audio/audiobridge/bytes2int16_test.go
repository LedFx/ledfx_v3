package audiobridge

import (
	"bytes"
	"ledfx/audio"
	"testing"
	"unsafe"
)

var (
	testBytes = bytes.Repeat([]byte("abc123"), 8192<<12)
)

func BenchmarkBytesToInt16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = bytesToAudioBuffer(testBytes)
	}
}

func BenchmarkBytesToInt16Unsafe(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = bytesToAudioBufferUnsafe(testBytes)
	}
}

func bytesToAudioBufferUnsafe(p []byte) (out audio.Buffer) {
	out = make([]int16, len(p))
	var offset int
	for i := 0; i < len(p); i += 2 {
		out[offset] = twoBytesToInt16Unsafe(p[i : i+2])
		offset++
	}
	return
}

func bytesToAudioBuffer(p []byte) (out audio.Buffer) {
	out = make([]int16, len(p))
	var offset int
	for i := 0; i < len(p); i += 2 {
		out[offset] = twoBytesToInt16(p[i : i+2])
		offset++
	}
	return
}

func twoBytesToInt16(p []byte) (out int16) {
	out |= int16(p[0])
	out |= int16(p[1]) << 8
	return
}

func twoBytesToInt16Unsafe(p []byte) (out int16) {
	return *(*int16)(unsafe.Pointer(&p[0]))
}
