package airplay2

import (
	"ledfx/handlers/raop"
	"testing"
)

type TestStruct struct {
	ptrVal  *raop.AirplayServer
	boolVal bool
}

func BenchmarkCheckNil_IsNil(b *testing.B) {
	ts := &TestStruct{
		ptrVal: nil,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for its := 0; its < 250000; its++ {
			if ts.ptrVal == nil {
				continue
			} else {
				b.Fatalf("ts.ptrVal should be nil!")
			}
		}
	}
}

func BenchmarkCheckNil_NotNil(b *testing.B) {
	ts := &TestStruct{
		ptrVal: new(raop.AirplayServer),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for its := 0; its < 250000; its++ {
			if ts.ptrVal != nil {
				continue
			} else {
				b.Fatalf("ts.ptrVal should not be nil!")
			}
		}
	}
}

func BenchmarkCheckBool_false(b *testing.B) {
	ts := &TestStruct{
		boolVal: false,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for its := 0; its < 250000; its++ {
			if !ts.boolVal {
				continue
			} else {
				b.Fatalf("ts.boolVal should not be true!")
			}
		}
	}
}

func BenchmarkCheckBool_true(b *testing.B) {
	ts := &TestStruct{
		boolVal: true,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for its := 0; its < 250000; its++ {
			if ts.boolVal {
				continue
			} else {
				b.Fatalf("ts.boolVal should not be false!")
			}
		}
	}
}
