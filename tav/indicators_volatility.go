package tav

import "math"

// ATR 平均真实波动范围，对应 stateful 指标实现中的 ATR
func ATR(high, low, close []float64, period int) []float64 {
	tr := TR(high, low, close)
	return RMA(tr, period)
}

// BBANDS 布林带指标，对应 stateful 指标实现中的 BBANDS
func BBANDS(data []float64, period int, stdUp, stdDn float64) ([]float64, []float64, []float64) {
	n := len(data)
	upper := make([]float64, n)
	middle := make([]float64, n)
	lower := make([]float64, n)

	if n < period {
		for i := range upper {
			upper[i] = math.NaN()
			middle[i] = math.NaN()
			lower[i] = math.NaN()
		}
		return upper, middle, lower
	}

	// 计算标准差和均值
	stddevs, means := StdDevBy(data, period, 0)
	copy(middle, means)

	// 计算上下轨
	for i := 0; i < n; i++ {
		if !math.IsNaN(middle[i]) && !math.IsNaN(stddevs[i]) {
			upper[i] = middle[i] + stddevs[i]*stdUp
			lower[i] = middle[i] - stddevs[i]*stdDn
		} else {
			upper[i] = math.NaN()
			lower[i] = math.NaN()
		}
	}

	return upper, middle, lower
}

// CHOP 乔普指数，对应 stateful 指标实现中的 CHOP
func CHOP(high, low, close []float64, period int) []float64 {
	n := len(close)
	res := make([]float64, n)

	// 当周期小于等于1时，log10(period)会是无效值
	if period <= 1 {
		for i := range res {
			res[i] = math.NaN()
		}
		return res
	}

	// ATR周期固定为1
	atr1 := ATR(high, low, close, 1)

	// 计算所需中间指标
	atrSum := Sum(atr1, period)
	hh := Highest(high, period)
	ll := Lowest(low, period)

	logPeriod := math.Log10(float64(period))

	for i := 0; i < n; i++ {
		s := atrSum[i]
		h := hh[i]
		l := ll[i]

		// 计算高低点范围
		rng := h - l
		if math.IsNaN(s) || math.IsNaN(rng) || rng == 0 {
			res[i] = math.NaN()
			continue
		}

		ratio := s / rng
		// log10只接受正数输入
		if ratio <= 0 {
			res[i] = math.NaN()
			continue
		}

		res[i] = 100 * math.Log10(ratio) / logPeriod
	}

	return res
}

// Squeeze reports 1 when two-standard-deviation Bollinger Bands fit inside a
// 1.5-ATR Keltner Channel, and 0 otherwise.
func Squeeze(high, low, close []float64, period int) []float64 {
	r := make([]float64, len(close))
	for i := range r {
		r[i] = math.NaN()
	}
	if period < 1 {
		return r
	}
	for i := period - 1; i < len(close); i++ {
		var sum, trs float64
		valid := true
		for j := 0; j < period; j++ {
			k := i - j
			if math.IsNaN(close[k]) || math.IsNaN(high[k]) || math.IsNaN(low[k]) {
				valid = false
				break
			}
			sum += close[k]
			prev := close[k]
			if k > 0 {
				prev = close[k-1]
			}
			if k == 0 || math.IsNaN(prev) {
				prev = close[k]
			}
			trs += math.Max(high[k], prev) - math.Min(low[k], prev)
		}
		if !valid {
			continue
		}
		mean := sum / float64(period)
		var ss float64
		for j := 0; j < period; j++ {
			d := close[i-j] - mean
			ss += d * d
		}
		std := math.Sqrt(ss / float64(period))
		kc := 1.5 * trs / float64(period)
		r[i] = 0
		if mean+2*std <= mean+kc && mean-2*std >= mean-kc {
			r[i] = 1
		}
	}
	return r
}

// StdDev 标准差，对应 stateful 指标实现中的 StdDev
func StdDev(data []float64, period int) []float64 {
	stddev, _ := StdDevBy(data, period, 0)
	return stddev
}

// StdDevBy 带自由度的标准差
func StdDevBy(data []float64, period int, ddof int) ([]float64, []float64) {
	n := len(data)
	stddev := make([]float64, n)

	mean := SMA(data, period)

	for i := 0; i < n; i++ {
		meanVal := mean[i]
		if math.IsNaN(meanVal) {
			stddev[i] = math.NaN()
			continue
		}
		var sumSqrt float64
		validNum := 0
		for j := 0; j <= i && validNum < period; j++ {
			pVal := data[i-j]
			if math.IsNaN(pVal) {
				continue
			}
			validNum += 1
			diff := pVal - meanVal
			sumSqrt += diff * diff
		}
		if validNum < period {
			stddev[i] = math.NaN()
			continue
		}
		variance := sumSqrt / float64(period-ddof)
		stddev[i] = math.Sqrt(variance)
	}

	return stddev, mean
}

// Stiffness 并行计算版本
func Stiffness(data []float64, maLen, stiffLen, stiffMa int) []float64 {
	n := len(data)
	if n == 0 {
		return []float64{}
	}

	// Step 1: 计算边界值 (SMA - StdDev * 0.2)
	bounds := calculateBounds(data, maLen)

	// Step 2: 计算原始大于边界的值 (转换为百分比)
	rawGt := calculateRawGreaterThan(data, bounds, stiffLen)

	// Step 3: 计算滑动窗口和
	sumValues := Sum(rawGt, stiffLen)

	// Step 4: 应用EMA
	return EMA(sumValues, stiffMa)
}

// Supertrend returns the conventional ATR-band trend line.
func Supertrend(high, low, close []float64, period int, multiplier float64) []float64 {

	n := len(close)
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}

	atr := ATR(high, low, close, period)
	if period < 1 {
		return out
	}
	fu, fl := math.NaN(), math.NaN()
	up := true
	for i := 0; i < n; i++ {
		if math.IsNaN(atr[i]) {
			continue
		}
		mid := (high[i] + low[i]) / 2
		bu, bl := mid+multiplier*atr[i], mid-multiplier*atr[i]
		if math.IsNaN(fu) {
			fu, fl = bu, bl
			up = close[i] >= mid
		} else {
			if bu < fu || close[i-1] > fu {
				fu = bu
			}
			if bl > fl || close[i-1] < fl {
				fl = bl
			}
			if up && close[i] < fl {
				up = false
			} else if !up && close[i] > fu {
				up = true
			}
		}
		if up {
			out[i] = fl
		} else {
			out[i] = fu
		}
	}
	return out
}

// TR 真实波动范围，对应 stateful 指标实现中的 TR
func TR(high, low, close []float64) []float64 {
	n := len(high)
	res := make([]float64, n)

	c1 := math.NaN()
	for i := 0; i < n; i++ {
		h, l := high[i], low[i]
		res[i] = max(h-l, math.Abs(h-c1), math.Abs(l-c1))
		cc := close[i]
		if !math.IsNaN(cc) {
			c1 = cc
		}
	}

	return res
}

func NATR(high, low, close []float64, period int) []float64 {
	a := ATR(high, low, close, period)
	r := make([]float64, len(close))
	for i := range r {
		if math.IsNaN(a[i]) || close[i] == 0 {
			r[i] = math.NaN()
		} else {
			r[i] = 100 * a[i] / close[i]
		}
	}
	return r
}

// AvgDev 平均偏差，对应 stateful 指标实现中的 AvgDev
func AvgDev(data []float64, period int) []float64 {
	n := len(data)
	res := make([]float64, n)

	if n < period {
		for i := range res {
			res[i] = math.NaN()
		}
		return res
	}

	// 计算移动平均
	sma := SMA(data, period)

	// 计算平均偏差
	for i := period - 1; i < n; i++ {
		mean := sma[i]
		if math.IsNaN(mean) {
			res[i] = math.NaN()
			continue
		}

		sumDev := 0.0
		validNum := 0
		for j := 0; validNum < period && j <= i; j++ {
			val := data[i-j]
			if math.IsNaN(val) {
				continue
			}
			validNum++
			sumDev += math.Abs(val - mean)
		}

		res[i] = sumDev / float64(period)
	}

	// 填充前面的NaN
	for i := 0; i < period-1; i++ {
		res[i] = math.NaN()
	}

	return res
}
