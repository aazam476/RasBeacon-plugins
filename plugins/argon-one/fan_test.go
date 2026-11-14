package main

import (
	"bytes"
	"testing"
)

// The fan speed goes to the case as one byte, the percent, which is what
// the Argon ONE V2 obeys: never the register form (0x80 first), which it
// acknowledges but ignores.
func TestFanSpeedIsOneByte(t *testing.T) {
	for _, tc := range []struct {
		percent int
		want    []byte
	}{
		{0, []byte{0}},
		{55, []byte{55}},
		{100, []byte{100}},
		{140, []byte{100}}, // at most 100%
		{-5, []byte{0}},
	} {
		var sent bytes.Buffer
		if err := (&fanController{w: &sent}).set(tc.percent); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(sent.Bytes(), tc.want) {
			t.Errorf("set(%d) sent %v, want %v", tc.percent, sent.Bytes(), tc.want)
		}
	}
}
