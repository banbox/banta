package banta

import (
	"encoding/json"
	"os"
)

var DataKline = loadKlines()

func loadKlines() []Kline {
	b, err := os.ReadFile("testdata/fixture_btc_58.json")
	if err != nil {
		panic(err)
	}
	var candles [][]float64
	if err := json.Unmarshal(b, &candles); err != nil {
		panic(err)
	}
	result := make([]Kline, len(candles))
	for i, c := range candles {
		if len(c) < 6 {
			panic("fixture candle must have six values")
		}
		result[i] = Kline{Time: int64(c[0]), Open: c[1], High: c[2], Low: c[3], Close: c[4], Volume: c[5]}
	}
	return result
}

func RunFakeEnv(env *BarEnv, klines []Kline, barCb func(int, Kline)) {
	env.Reset()
	for i, bar := range klines {
		env.OnBar(bar.Time, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume, bar.Quote, bar.BuyVolume, bar.TradeNum)
		if barCb != nil {
			barCb(i, bar)
		}
	}
}
