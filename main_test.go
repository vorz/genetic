package main

import (
	"testing"
)

func TestMakeParseFile(t *testing.T) {
	var out1, out2 []xy
	out1 = makeFile(500, func(x float64) float64 {
		return x * 3
	})
	out2 = parseFile()
	if len(out1) != len(out2) {
		t.Errorf("Length of xy from making and parsing differs (%v vs %v)", len(out1), len(out2))
	}
	for n, v := range out1 {
		if v.x != out2[n].x || v.y != out2[n].y {
			t.Errorf("Values of xy differs at %d element: %v vs %v", n, v, out2[n])
		}
	}
}
