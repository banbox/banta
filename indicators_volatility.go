package banta

import "github.com/banbox/banta/tav"
import "math"

/*
	StdDev Standard Deviation 标准差

suggest period: 20
*/
func StdDev(obj *Series, period int) *Series {
	sd, _ := StdDevBy(obj, period, 0)
	return sd
}

/*
	StdDevBy Standard Deviation

suggest period: 20

return [stddev，sumVal]
*/
func StdDevBy(obj *Series, period int, ddof int) (*Series, *Series) {
	res := obj.To("_sdev", period*10+ddof)
	if !res.Cached() {
		if !res.Cached() {
			meanVal := SMA(obj, period).Get(0)
			inVal := obj.Get(0)
			if math.IsNaN(inVal) {
				res.Append([]float64{math.NaN(), math.NaN()})
			} else {
				arr := WrapFloatArr(res, period, inVal)
				if len(arr) < period {
					res.Append([]float64{math.NaN(), math.NaN()})
				} else {
					sumSqrt := 0.0
					for _, x := range arr {
						sumSqrt += (x - meanVal) * (x - meanVal)
					}
					variance := sumSqrt / float64(period-ddof)
					stdDevVal := math.Sqrt(variance)
					res.Append([]float64{stdDevVal, meanVal})
				}
			}
		}
	}
	return res, res.Cols[0]
}

/*
ATR Average True Range

suggest period: 14
*/
func ATR(high *Series, low *Series, close *Series, period int) *Series {
	return RMA(TR(high, low, close), period)
}

/*
BBANDS Bollinger Bands 布林带指标

period: 20, stdUp: 2, stdDn: 2

return [upper, mid, lower]
*/
func BBANDS(obj *Series, period int, stdUp, stdDn float64) (*Series, *Series, *Series) {
	res := obj.To("_bb", pkey(ikey(period), fkey(stdUp), fkey(stdDn)))
	if !res.Cached() {
		if !res.Cached() {
			devCol, meanCol := StdDevBy(obj, period, 0)
			dev, mean := devCol.Get(0), meanCol.Get(0)
			if math.IsNaN(dev) {
				res.Append([]float64{math.NaN(), math.NaN(), math.NaN()})
			} else {
				upper := mean + dev*stdUp
				lower := mean - dev*stdDn

				res.Append([]float64{upper, mean, lower})
			}
		}
	}
	return res, res.Cols[0], res.Cols[1]
}

/*
CHOP Choppiness Index

suggest period: 14

higher values equal more choppiness, while lower values indicate directional trending.
值越高，波动性越大，而值越低，则表示有方向性趋势。
*/
func CHOP(e *BarEnv, period int) *Series {
	res := e.Close.To("_chop", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		atrSum := Sum(ATR(e.High, e.Low, e.Close, 1), period).Get(0)
		hh := Highest(e.High, period).Get(0)
		ll := Lowest(e.Low, period).Get(0)
		val := 100 * math.Log10(atrSum/(hh-ll)) / math.Log10(float64(period))
		res.Append(val)
	}
	return res
}

/*
Stiffness Indicator

maLen: 100, stiffLen: 60, stiffMa: 3
*/
func Stiffness(obj *Series, maLen, stiffLen, stiffMa int) *Series {
	bound := obj.To("_sti_bound", maLen)
	if !bound.Cached() {
		if !bound.Cached() {
			stdDev := StdDev(obj, maLen)
			bound.Append(SMA(obj, maLen).Get(0) - stdDev.Get(0)*0.2)
		}
	}
	above := bound.To("_raw_gt", stiffLen)
	if !above.Cached() {
		if !above.Cached() {
			boundVal := bound.Get(0)
			if math.IsNaN(boundVal) {
				above.Append(math.NaN())
			} else {
				val := float64(0)
				if obj.Get(0) > boundVal {
					val = 100 / float64(stiffLen)
				}
				above.Append(val)
			}
		}
	}
	return EMA(Sum(above, stiffLen), stiffMa)
}

// STDDEV preserves the uppercase spelling emitted by some source adapters.
func STDDEV(obj *Series, period int) *Series { return StdDev(obj, period) }

// Squeeze reports the TTM squeeze condition as 1 when Bollinger Bands are
// inside a Keltner Channel, and 0 otherwise. Both use the same period;
// BB uses two standard deviations and KC uses 1.5 ATR.
func Squeeze(high, low, close *Series, period int) *Series {
	res := close.To("_squeeze", period)
	if res.Cached() {
		return res
	}
	if period < 1 || close.Len() < period {
		res.Append(math.NaN())
		return res
	}
	var sum, trSum float64
	for i := 0; i < period; i++ {
		c, h, l := close.Get(i), high.Get(i), low.Get(i)
		if math.IsNaN(c) || math.IsNaN(h) || math.IsNaN(l) {
			res.Append(math.NaN())
			return res
		}
		sum += c
		prev := close.Get(i + 1)
		if math.IsNaN(prev) {
			prev = c
		}
		trSum += math.Max(h, prev) - math.Min(l, prev)
	}
	mean := sum / float64(period)
	var ss float64
	for i := 0; i < period; i++ {
		d := close.Get(i) - mean
		ss += d * d
	}
	std := math.Sqrt(ss / float64(period))
	bbUp, bbDn := mean+2*std, mean-2*std
	kc := mean + 1.5*(trSum/float64(period))
	if bbUp <= kc && bbDn >= mean-1.5*(trSum/float64(period)) {
		res.Append(1.0)
	} else {
		res.Append(0.0)
	}
	return res
}

func NATR(high, low, close *Series, period int) *Series {
	res := close.To("_natr", period)
	if !res.Cached() {
		atr := ATR(high, low, close, period).Get(0)
		c := close.Get(0)
		if math.IsNaN(atr) || c == 0 {
			res.Append(math.NaN())
		} else {
			res.Append(100 * atr / c)
		}
	}
	return res
}

func Supertrend(high, low, close *Series, period int, multiplier float64) *Series {
	r := close.To("_supertrend", pkey(ikey(period), fkey(multiplier)))
	if r.Cached() {
		return r
	}
	v := tav.Supertrend(hist(high), hist(low), hist(close), period, multiplier)
	r.Append(last(v))
	return r
}

func TR(high *Series, low *Series, close *Series) *Series {
	res := high.To("_tr", 0)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		pclose, ok := res.More.(float64)
		if !ok {
			pclose = math.NaN()
		}
		resVal := math.NaN()
		if high.Len() >= 2 {
			chigh, clow := high.Get(0), low.Get(0)
			resVal = max(chigh-clow, math.Abs(chigh-pclose), math.Abs(clow-pclose))
		}
		curClose := close.Get(0)
		if !math.IsNaN(curClose) {
			res.More = curClose
		}
		res.Append(resVal)
	}
	return res
}

// AvgDev sum(abs(Vi - mean))/period
func AvgDev(obj *Series, period int) *Series {
	res := obj.To("_avgdev", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		sma := SMA(obj, period)
		smaVal := sma.Get(0)
		inVal := obj.Get(0)
		if math.IsNaN(smaVal) || math.IsNaN(inVal) {
			res.Append(math.NaN())
		} else {
			sumDev := 0.0
			validNum := 0
			for i := 0; validNum < period; i++ {
				val := obj.Get(i)
				if math.IsNaN(val) {
					continue
				}
				validNum++
				sumDev += math.Abs(val - smaVal)
			}

			avgDev := sumDev / float64(period)
			res.Append(avgDev)
		}
	}
	return res
}
