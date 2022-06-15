package util

import (
	"testing"
)

// This is the length of the current embedded logo.txt file.
// We assume that nothing else within the logo.go file will ever break.
var logoLength = 60276

func TestLogoSize(t *testing.T) {
	if len(logoTxt) != logoLength {
		t.Errorf("Logo length has changed - did you update the logo? %d != %d", len(logoTxt), logoLength)
	}
}
