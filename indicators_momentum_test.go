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

func TestRMIAllowsMultiplePeriodsWithSharedMomentumLength(t *testing.T) {
	e, err := NewBarEnv("momentum", "spot", "BTC/USDT", "1d")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RMI panicked when multiple periods shared montLen: %v", r)
		}
	}()
	for i := 0; i < 80; i++ {
		close := 100 + float64((i*7)%23) + float64(i)/10
		if err := e.OnBar(int64(i+1)*86400000, close-1, close+2, close-2, close, 10, 0, 0, 0); err != nil {
			t.Fatal(err)
		}
		_ = RMI(e.Close, 8, 3).Get(0)
		_ = RMI(e.Close, 12, 3).Get(0)
		_ = RMI(e.Close, 30, 3).Get(0)
		_ = RMI(e.Close, 520, 3).Get(0)
	}
}
