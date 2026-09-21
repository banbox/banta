package banta

import (
	"math"
	"testing"

	"github.com/banbox/banta/tav"
)

func batchCData() (close, volume, high, low []float64) {
	close = []float64{10, 11, 13, 12, 14, 15, 14, 16, 18, 17, 19, 20, 21, 20, 22, 24}
	volume = []float64{2, 3, 4, 2, 5, 3, 4, 6, 2, 3, 5, 4, 3, 6, 2, 5}
	high, low = make([]float64, len(close)), make([]float64, len(close))
	for i, c := range close {
		high[i], low[i] = c+1, c-1
	}
	return
}

func TestBatchCVectorDefinitions(t *testing.T) {
	close, volume, high, low := batchCData()
	assertFloatSliceParity(t, "EFI", tav.EFI(close, volume, 3), tav.EMA(func() []float64 {
		r := make([]float64, len(close))
		for i := range r {
			r[i] = math.NaN()
		}
		for i := 1; i < len(r); i++ {
			r[i] = (close[i] - close[i-1]) * volume[i]
		}
		return r
	}(), 3), 1e-10)
	u, m, d := tav.Donchian(high, low, 4)
	for i := 3; i < len(close); i++ {
		expU, expD := high[i], low[i]
		for j := i - 3; j <= i; j++ {
			if high[j] > expU {
				expU = high[j]
			}
			if low[j] < expD {
				expD = low[j]
			}
		}
		if u[i] != expU || d[i] != expD || m[i] != (u[i]+d[i])/2 {
			t.Fatalf("donchian bar %d: %v %v %v", i, u[i], m[i], d[i])
		}
	}
	if !math.IsNaN(tav.Squeeze(high, low, close, 4)[0]) {
		t.Fatal("squeeze warmup missing")
	}
	assertFloatSliceParity(t, "Slope", tav.Slope(close, 4), tav.LinRegAdv(close, 4, false, false, false, false, true, false), 1e-10)
}

func TestBatchCStateVectorParity(t *testing.T) {
	close, volume, high, low := batchCData()
	ks := make([]Kline, len(close))
	for i := range ks {
		ks[i] = Kline{Time: int64(i + 1), Open: close[i], High: high[i], Low: low[i], Close: close[i], Volume: volume[i]}
	}
	e, err := NewBarEnv("test", "spot", "BTC/USDT", "1d")
	if err != nil {
		t.Fatal(err)
	}
	var efi, sq, slope, du, dm, dd []float64
	RunFakeEnv(e, ks, func(_ int, _ Kline) {
		efi = append(efi, EFI(e, 3).Get(0))
		sq = append(sq, Squeeze(e.High, e.Low, e.Close, 4).Get(0))
		slope = append(slope, Slope(e.Close, 4).Get(0))
		u, m, d := Donchian(e.High, e.Low, 4)
		du = append(du, u.Get(0))
		dm = append(dm, m.Get(0))
		dd = append(dd, d.Get(0))
	})
	assertFloatSliceParity(t, "EFI", efi, tav.EFI(close, volume, 3), 1e-10)
	assertFloatSliceParity(t, "Squeeze", sq, tav.Squeeze(high, low, close, 4), 1e-10)
	assertFloatSliceParity(t, "Slope", slope, tav.Slope(close, 4), 1e-10)
	u, m, d := tav.Donchian(high, low, 4)
	assertFloatSliceParity(t, "Donchian upper", du, u, 1e-10)
	assertFloatSliceParity(t, "Donchian middle", dm, m, 1e-10)
	assertFloatSliceParity(t, "Donchian lower", dd, d, 1e-10)
}
