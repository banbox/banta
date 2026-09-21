package banta

import (
	"math"
	"testing"

	"github.com/banbox/banta/tav"
)

func batchBFixture() []float64 {
	return []float64{10, 11, 13, 12, 14, 15, 16, 18, 17, 19, 20, 21, 23, 22, 24, 25, 26, 28, 27, 29}
}

func TestBatchBVectorParity(t *testing.T) {
	data := batchBFixture()
	for _, period := range []int{1, 5, 10} {
		wantAngle := tav.LINEARREG_ANGLE(data, period)
		wantDPO := tav.DPO(data, period)
		gotAngle := tav.LinearRegAngle(data, period)
		gotDPO := tav.Dpo(data, period)
		assertFloatSliceParity(t, "LINEARREG_ANGLE", gotAngle, wantAngle, 1e-12)
		assertFloatSliceParity(t, "DPO", gotDPO, wantDPO, 1e-12)
	}
}

func TestBatchBStateVectorParity(t *testing.T) {
	data := batchBFixture()
	klines := make([]Kline, len(data))
	for i, v := range data {
		klines[i] = Kline{Time: int64(i+1) * 86400000, Open: v, High: v + 1, Low: v - 1, Close: v, Volume: 1}
	}
	for _, period := range []int{1, 5, 10} {
		env, err := NewBarEnv("test", "spot", "BTC/USDT", "1d")
		if err != nil {
			t.Fatal(err)
		}
		var angle, dpo []float64
		RunFakeEnv(env, klines, func(_ int, _ Kline) {
			angle = append(angle, LinearRegAngle(env.Close, period).Get(0))
			dpo = append(dpo, DPO(env.Close, period).Get(0))
		})
		assertFloatSliceParity(t, "LINEARREG_ANGLE", angle, tav.LINEARREG_ANGLE(data, period), 1e-12)
		assertFloatSliceParity(t, "DPO", dpo, tav.DPO(data, period), 1e-12)
	}
}

func TestBatchBIndependentFormula(t *testing.T) {
	data := []float64{2, 4, 6, 8, 10, 12, 14}
	got := tav.DPO(data, 3)
	// centered=False uses close[t] - SMA[t-shift], shift=2; first value is bar 4.
	if !math.IsNaN(got[0]) || !math.IsNaN(got[3]) || got[4] != 6 || got[5] != 6 {
		t.Fatalf("unexpected DPO warmup/flat values: %#v", got)
	}
	angle := tav.LINEARREG_ANGLE(data, 3)
	for i := 0; i < 2; i++ {
		if !math.IsNaN(angle[i]) {
			t.Fatalf("angle warmup bar %d = %v", i, angle[i])
		}
	}
	for i := 2; i < len(angle); i++ {
		if math.Abs(angle[i]-math.Atan(2)*180/math.Pi) > 1e-14 {
			t.Fatalf("angle bar %d = %.17g", i, angle[i])
		}
	}
}

func TestBatchBExternalOracles(t *testing.T) {
	for _, name := range []string{"linearreg_angle", "dpo"} {
		f := loadCompareFixture(t, name)
		data := make([]float64, len(f.Candles))
		for i, c := range f.Candles {
			data[i] = c[4]
		}
		if name == "linearreg_angle" {
			compareExternal(t, name, "talib", tav.LINEARREG_ANGLE(data, 14))
		} else {
			compareExternal(t, name, "pandas_ta", tav.DPO(data, 20))
		}
	}
}
