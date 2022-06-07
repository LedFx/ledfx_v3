package audio

import (
	"testing"
)

var testAudio Buffer = make(Buffer, framesPerBuffer*4)

func BenchmarkAnalysis(t *testing.B) {
	for i := 0; i < t.N; i++ {
		Analyzer.BufferCallback(testAudio)
	}
}
