package banta

import (
	"math"
	"testing"

	"github.com/banbox/banta/tav"
)

func TestBatchEStateVectorParity(t *testing.T) {
	ks := extraFixture()
	o, h, l, c, _, _ := extractOHLCV(ks)
	_ = o
	e, err := NewBarEnv("batch-e", "spot", "", "1d")
	if err != nil {
		t.Fatal(err)
	}
	var st, vi, sf, pm, pd, ku, km, kl []float64
	RunFakeEnv(e, ks, func(_ int, _ Kline) {
		st = append(st, Supertrend(e.High, e.Low, e.Close, 3, 2).Get(0))
		vi = append(vi, VIDYA(e.Close, 3).Get(0))
		sf = append(sf, SSF(e.Close, 3).Get(0))
		p, d := PMAX(e.High, e.Low, e.Close, 3, 2)
		pm = append(pm, p.Get(0))
		pd = append(pd, d.Get(0))
		u, m, d2 := KeltnerChannel(e.High, e.Low, e.Close, 3, 2)
		ku = append(ku, u.Get(0))
		km = append(km, m.Get(0))
		kl = append(kl, d2.Get(0))
	})
	assertStateVectorParity(t, "Supertrend", tav.Supertrend(h, l, c, 3, 2), func(e *BarEnv) float64 { return Supertrend(e.High, e.Low, e.Close, 3, 2).Get(0) })
	assertStateVectorParity(t, "VIDYA", tav.VIDYA(c, 3), func(e *BarEnv) float64 { return VIDYA(e.Close, 3).Get(0) })
	assertStateVectorParity(t, "SSF", tav.SSF(c, 3), func(e *BarEnv) float64 { return SSF(e.Close, 3).Get(0) })
	vpm, vpd := tav.PMAX(h, l, c, 3, 2)
	for name, got := range map[string][]float64{"PMAX": pm, "PMAX direction": pd, "KC upper": ku, "KC middle": km, "KC lower": kl} {
		var ref []float64
		switch name {
		case "PMAX":
			ref = vpm
		case "PMAX direction":
			ref = vpd
		case "KC upper", "KC middle", "KC lower":
			a, b, d := tav.KeltnerChannel(h, l, c, 3, 2)
			if name == "KC upper" {
				ref = a
			} else if name == "KC middle" {
				ref = b
			} else {
				ref = d
			}
		}
		if len(got) != len(ref) {
			t.Fatalf("%s length", name)
		}
		for i := range got {
			if math.IsNaN(got[i]) && math.IsNaN(ref[i]) {
				continue
			}
			if math.Abs(got[i]-ref[i]) > 1e-9 {
				t.Fatalf("%s[%d] %v != %v", name, i, got[i], ref[i])
			}
		}
	}
}
