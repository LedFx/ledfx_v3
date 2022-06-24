package audio

import (
	"testing"
)

var testAudio Buffer = make(Buffer, framesPerBuffer*2)

func TestAnalysis(t *testing.T) {
	Analyzer.NewMelbank("totally_valid_id", Mono, uint(20), uint(20000), 0.7)
	Analyzer.BufferCallback(testAudio)
	Analyzer.Cleanup()
}

func BenchmarkAnalysis(t *testing.B) {
	Analyzer.NewMelbank("totally_valid_id", Mono, uint(20), uint(20000), 0.7)
	for i := 0; i < t.N; i++ {
		Analyzer.BufferCallback(testAudio)
	}
	Analyzer.Cleanup()
}
