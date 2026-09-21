package banta

import (
	"github.com/banbox/banta/tav"
	"math"
	"testing"
)

func TestClassicIndicatorsStateVectorParity(t *testing.T) {
	ks := extraFixture()
	_, h, l, c, v, _ := extractOHLCV(ks)
	check := func(name string, want []float64, run func(*BarEnv) float64) {
		e, _ := NewBarEnv("classic", "spot", "", "1d")
		got := []float64{}
		RunFakeEnv(e, ks, func(_ int, _ Kline) { got = append(got, run(e)) })
		if len(got) != len(want) {
			t.Fatalf("%s length", name)
		}
		for i := range got {
			if math.IsNaN(got[i]) && math.IsNaN(want[i]) {
				continue
			}
			if math.Abs(got[i]-want[i]) > 1e-8 {
				t.Fatalf("%s[%d] %v != %v", name, i, got[i], want[i])
			}
		}
	}
	check("KST", tav.KST(c, 2, 3, 4, 5, 2, 2, 2, 2), func(e *BarEnv) float64 { return KST(e.Close, 2, 3, 4, 5, 2, 2, 2, 2).Get(0) })
	check("PBand", tav.DonchianPBand(h, l, c, 3), func(e *BarEnv) float64 { return DonchianPBand(e.High, e.Low, e.Close, 3).Get(0) })
	check("WBand", tav.KeltnerWBand(h, l, c, 3, 2), func(e *BarEnv) float64 { return KeltnerWBand(e.High, e.Low, e.Close, 3, 2).Get(0) })
	check("VPCI", tav.VPCI(c, v, 3), func(e *BarEnv) float64 { return VPCI(e.Close, e.Volume, 3).Get(0) })
	check("Williams", tav.WilliamsPercent(h, l, c, 3), func(e *BarEnv) float64 { return WilliamsPercent(e.High, e.Low, e.Close, 3).Get(0) })
	check("DX", tav.DX(h, l, c, 3), func(e *BarEnv) float64 { return DX(e.High, e.Low, e.Close, 3).Get(0) })
	check("Fisher", tav.Fisher(h, l, 3), func(e *BarEnv) float64 { return Fisher(e.High, e.Low, 3).Get(0) })
	check("Correlation", tav.Correlation(c, v, 3), func(e *BarEnv) float64 { return Correlation(e.Close, e.Volume, 3).Get(0) })
	e1, b1, a1, s1, lag1 := tav.Ichimoku(h, l, c, 2, 3, 4)
	e, _ := NewBarEnv("ich", "spot", "", "1d")
	var got [5][]float64
	RunFakeEnv(e, ks, func(_ int, _ Kline) {
		x, y, z, q, w := Ichimoku(e.High, e.Low, e.Close, 2, 3, 4)
		got[0] = append(got[0], x.Get(0))
		got[1] = append(got[1], y.Get(0))
		got[2] = append(got[2], z.Get(0))
		got[3] = append(got[3], q.Get(0))
		got[4] = append(got[4], w.Get(0))
	})
	for j, want := range [5][]float64{e1, b1, a1, s1, lag1} {
		for i := range want {
			if math.IsNaN(got[j][i]) && math.IsNaN(want[i]) {
				continue
			}
			if math.Abs(got[j][i]-want[i]) > 1e-8 {
				t.Fatalf("Ichimoku line %d bar %d", j, i)
			}
		}
	}
	mama, fama := tav.MAMA(c, 2, 30)
	check("MAMA", mama, func(e *BarEnv) float64 { x, _ := MAMA(e.Close, 2, 30); return x.Get(0) })
	_ = fama
}
