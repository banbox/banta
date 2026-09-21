package banta

import (
	"math"
	"testing"
)

func TestGeneratedIndicatorCompatibility(t *testing.T) {
	env, err := NewBarEnv("test", "spot", "BTC/USDT", "1m")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 80; i++ {
		close := 100 + float64(i%11) + float64(i)/20
		env.OnBar(int64(i+1)*60000, close-1, close+2, close-2, close, 10+float64(i%5), 0, 0, 0)
		if env.BarNum > 20 {
			plus, minus, adx := DMI(env.High, env.Low, env.Close, 14, 14)
			if plus == nil || minus == nil || adx == nil {
				t.Fatal("DMI returned nil series")
			}
			if got := STOCH(env.Close, env.High, env.Low, 14); got == nil {
				t.Fatal("STOCH returned nil series")
			}
			if got := VWAP(env.High, env.Low, env.Close, env.Volume); got == nil {
				t.Fatal("VWAP returned nil series")
			} else if !math.IsNaN(got.Get(0)) && got.Get(0) <= 0 {
				t.Fatalf("VWAP=%v", got.Get(0))
			}
		}
	}
}
