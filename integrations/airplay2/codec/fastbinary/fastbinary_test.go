package fastbinary

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestInternalBinary_Read(t *testing.T) {
	inbuf := []byte("gg")
	outbuf := int16(0)
	if err := binary.Read(bytes.NewReader(inbuf), binary.LittleEndian, &outbuf); err != nil {
		t.Fatalf("Error reading into buffer: %v\n", err)
	}
	if outbuf != 26471 {
		t.Fatalf("Expected: 26471\nGot: %d\n", outbuf)
	}
}

func TestFastBinary_Read(t *testing.T) {
	inbuf := []byte("gg")
	outbuf := int16(0)
	if err := ReadInt16(bytes.NewReader(inbuf), &outbuf); err != nil {
		t.Fatalf("Error reading into buffer: %v\n", err)
	}
	if outbuf != 26471 {
		t.Fatalf("Expected: 26471\nGot: %d\n", outbuf)
	}
}

func TestFastBinaryNoIoReader_Read(t *testing.T) {
	inbuf := []byte("gg")
	outbuf := int16(0)
	ReadInt16FromBytes(inbuf, &outbuf)
	if outbuf != 26471 {
		t.Fatalf("Expected: 26471\nGot: %d\n", outbuf)
	}
}

func BenchmarkInternalBinary_Read(b *testing.B) {
	inbuf := []byte("gg")
	for i := 0; i < b.N; i++ {
		for rds := 0; rds < 10000; rds++ {
			outbuf := int16(0)
			if err := binary.Read(bytes.NewReader(inbuf), binary.LittleEndian, &outbuf); err != nil {
				b.Fatalf("Error reading into buffer: %v\n", err)
			}
		}
	}
}

func BenchmarkFastBinary_Read(b *testing.B) {
	inbuf := []byte("gg")
	for i := 0; i < b.N; i++ {
		for rds := 0; rds < 10000; rds++ {
			outbuf := int16(0)
			if err := ReadInt16(bytes.NewReader(inbuf), &outbuf); err != nil {
				b.Fatalf("Error reading into buffer: %v\n", err)
			}
		}
	}
}

func BenchmarkFastBinaryNoIoReader_Read(b *testing.B) {
	inbuf := []byte("gg")
	for i := 0; i < b.N; i++ {
		for rds := 0; rds < 10000; rds++ {
			outbuf := int16(0)
			ReadInt16FromBytes(inbuf, &outbuf)
		}
	}
}
