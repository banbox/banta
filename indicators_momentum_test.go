package banta

import (
	"github.com/banbox/banta/tav"
	"math"
	"testing"
)

func TestMomentumIndicatorsParity(t *testing.T) {
	ks := extraFixture()
	_, _, _, c, _, _ := extractOHLCV(ks)
	e, _ := NewBarEnv("momentum", "spot", "", "1d")
	var tr, ts []float64
	RunFakeEnv(e, ks, func(_ int, _ Kline) {
		tr = append(tr, TRIX(e.Close, 2).Get(0))
		ts = append(ts, TSI(e.Close, 2, 3).Get(0))
	})
	for name, v := range map[string]struct{ g, w []float64 }{"TRIX": {tr, tav.TRIX(c, 2)}, "TSI": {ts, tav.TSI(c, 2, 3)}} {
		for i := range v.g {
			if math.IsNaN(v.g[i]) && math.IsNaN(v.w[i]) {
				continue
			}
			if math.Abs(v.g[i]-v.w[i]) > 1e-8 {
				t.Fatalf("%s[%d] %v != %v", name, i, v.g[i], v.w[i])
			}
		}
	}
}
