package banta

import "testing"

func benchmarkIndicatorStream(b *testing.B, run func(*BarEnv)) {
	e, err := NewBarEnv("bench", "spot", "BTC/USDT", "1m")
	if err != nil {
		b.Fatal(err)
	}
	e.MaxCache = 2000
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		price := 100 + float64(i%97)
		if err := e.OnBar(int64(i+1)*60000, price-1, price+2, price-2, price, 10, 1000, 5, 1); err != nil {
			b.Fatal(err)
		}
		run(e)
	}
}

func BenchmarkStreamMACD(b *testing.B) {
	benchmarkIndicatorStream(b, func(e *BarEnv) { MACD(e.Close, 12, 26, 9) })
}

func BenchmarkStreamALMA(b *testing.B) {
	benchmarkIndicatorStream(b, func(e *BarEnv) { ALMA(e.Close, 10, 6, 0.85) })
}

func BenchmarkStreamBBANDS(b *testing.B) {
	benchmarkIndicatorStream(b, func(e *BarEnv) { BBANDS(e.Close, 20, 2, 2) })
}

func BenchmarkStreamSTC(b *testing.B) {
	benchmarkIndicatorStream(b, func(e *BarEnv) { STC(e.Close, 12, 26, 50, 0.5) })
}

func BenchmarkStreamSAR(b *testing.B) {
	benchmarkIndicatorStream(b, func(e *BarEnv) { SAR(e.High, e.Low, 0.02, 0.2) })
}
