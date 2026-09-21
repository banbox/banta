package tav

import (
	"math"
	"testing"
)

func TestExtraIndicators(t *testing.T) {
	c := []float64{1, 2, 3, 2, 4, 5, 4, 6}
	h := []float64{2, 3, 4, 3, 5, 6, 5, 7}
	l := []float64{0, 1, 2, 1, 3, 4, 3, 5}
	v := []float64{1, 1, 1, 1, 1, 1, 1, 1}
	if got := MOM(c, 2); !equal(got[2], 2) {
		t.Fatalf("MOM=%v", got[2])
	}
	if got := OBV(c, v); got[len(got)-1] == 0 {
		t.Fatal("OBV did not accumulate")
	}
	if len(DEMA(c, 3)) != len(c) || len(T3(c, 3)) != len(c) {
		t.Fatal("moving average lengths")
	}
	if len(SAR(h, l, .02, .2)) != len(c) {
		t.Fatal("SAR length")
	}
	if !math.IsNaN(ULTOSC(h, l, c, 2, 3, 4)[0]) {
		t.Fatal("ULTOSC warmup")
	}
}
func equal(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
