package banta

import (
	"github.com/banbox/banta/tav"
	"math"
	"testing"
)

func TestCandlePatternStateVectorParity(t *testing.T) {
	ks := extraFixture()
	o, h, l, c, _, _ := extractOHLCV(ks)
	e, err := NewBarEnv("candles", "spot", "", "1d")
	if err != nil {
		t.Fatal(err)
	}
	patterns := []struct {
		name  string
		vec   []float64
		state func(*BarEnv) float64
	}{
		{"HAMMER", tav.CDLHAMMER(o, h, l, c), func(e *BarEnv) float64 { return CDLHAMMER(e.Open, e.High, e.Low, e.Close).Get(0) }},
		{"ENGULFING", tav.CDLENGULFING(o, h, l, c), func(e *BarEnv) float64 { return CDLENGULFING(e.Open, e.High, e.Low, e.Close).Get(0) }},
		{"3INSIDE", tav.CDL3INSIDE(o, h, l, c), func(e *BarEnv) float64 { return CDL3INSIDE(e.Open, e.High, e.Low, e.Close).Get(0) }},
		{"3OUTSIDE", tav.CDL3OUTSIDE(o, h, l, c), func(e *BarEnv) float64 { return CDL3OUTSIDE(e.Open, e.High, e.Low, e.Close).Get(0) }},
		{"3LINESTRIKE", tav.CDL3LINESTRIKE(o, h, l, c), func(e *BarEnv) float64 { return CDL3LINESTRIKE(e.Open, e.High, e.Low, e.Close).Get(0) }},
		{"DRAGONFLY", tav.CDLDRAGONFLYDOJI(o, h, l, c), func(e *BarEnv) float64 { return CDLDRAGONFLYDOJI(e.Open, e.High, e.Low, e.Close).Get(0) }},
		{"GRAVESTONE", tav.CDLGRAVESTONEDOJI(o, h, l, c), func(e *BarEnv) float64 { return CDLGRAVESTONEDOJI(e.Open, e.High, e.Low, e.Close).Get(0) }},
		{"HANGINGMAN", tav.CDLHANGINGMAN(o, h, l, c), func(e *BarEnv) float64 { return CDLHANGINGMAN(e.Open, e.High, e.Low, e.Close).Get(0) }},
		{"SHOOTINGSTAR", tav.CDLSHOOTINGSTAR(o, h, l, c), func(e *BarEnv) float64 { return CDLSHOOTINGSTAR(e.Open, e.High, e.Low, e.Close).Get(0) }},
		{"MORNINGSTAR", tav.CDLMORNINGSTAR(o, h, l, c), func(e *BarEnv) float64 { return CDLMORNINGSTAR(e.Open, e.High, e.Low, e.Close).Get(0) }},
	}
	for _, p := range patterns {
		e, _ = NewBarEnv("candles", "spot", "", "1d")
		got := []float64{}
		RunFakeEnv(e, ks, func(_ int, _ Kline) { got = append(got, p.state(e)) })
		if len(got) != len(p.vec) {
			t.Fatalf("%s length", p.name)
		}
		for i := range got {
			if math.IsNaN(got[i]) && math.IsNaN(p.vec[i]) {
				continue
			}
			if got[i] != p.vec[i] {
				t.Fatalf("%s[%d] %v != %v", p.name, i, got[i], p.vec[i])
			}
		}
	}
}
