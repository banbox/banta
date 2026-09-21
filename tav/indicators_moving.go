package tav

import "math"

// ALMA Arnaud Legoux Moving Average
/*
period:  window size.
sigma:   Smoothing value.
distOff: min 0 (smoother), max 1 (more responsive).
*/
func ALMA(data []float64, period int, sigma, distOff float64) []float64 {
	n := len(data)
	result := make([]float64, n)

	// Handle edge cases where calculation is not possible
	if period <= 0 {
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	// Pre-calculate weights and their sum
	m := distOff * (float64(period) - 1)
	s := float64(period) / sigma
	weights := make([]float64, period)
	cumSum := 0.0
	for i := 0; i < period; i++ {
		fi := float64(i)
		// Calculate weight for each position in the window
		w := math.Exp(-(fi - m) * (fi - m) / (2 * s * s))
		weights[i] = w
		cumSum += w
	}

	// Avoid division by zero if all weights are somehow zero
	if cumSum == 0 {
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	// Main calculation loop
	arr := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		// Not enough data points for a full window
		val := data[i]
		if math.IsNaN(val) {
			result[i] = math.NaN()
			continue
		}
		arr = append(arr, val)
		if len(arr) > period {
			arr = arr[1:]
		} else if len(arr) < period {
			result[i] = math.NaN()
			continue
		}

		windowSum := 0.0
		for j := 0; j < period; j++ {
			val = arr[period-j-1]
			windowSum += val * weights[j]
		}

		result[i] = windowSum / cumSum
	}

	return result
}

// EMA 指数移动平均线，对应 stateful 指标实现中的 EMA
func EMA(data []float64, period int) []float64 {
	return EMABy(data, period, 0)
}

// EMABy 指数移动平均线，可指定初始化方式
func EMABy(data []float64, period int, initType int) []float64 {
	alpha := 2.0 / float64(period+1)
	return ewma(data, period, alpha, initType, math.NaN())
}

// HMA an implementation of Hull Moving Average.
// This is a high-performance version for batch calculation.
func HMA(data []float64, period int) []float64 {
	n := len(data)
	// Return a slice of NaNs if period is invalid.
	if period <= 1 {
		res := make([]float64, n)
		for i := range res {
			res[i] = math.NaN()
		}
		return res
	}

	// Calculate the required periods for the HMA formula.
	periodSqrt := int(math.Floor(math.Sqrt(float64(period))))

	// Calculate the two initial WMAs.
	wmaHalf := WMA(data, period/2)
	wmaFull := WMA(data, period)

	// Calculate the intermediate series (2*WMA(half) - WMA(full)).
	diff := make([]float64, n)
	for i := 0; i < n; i++ {
		diff[i] = 2*wmaHalf[i] - wmaFull[i]
	}

	// The final result is a WMA of the intermediate series.
	return WMA(diff, periodSqrt)
}

// KAMA Kaufman Adaptive Moving Average 并行计算版本
// period: 10 fixed: (fast: 2, slow: 30)
func KAMA(data []float64, period int) []float64 {
	return KAMABy(data, period, 2, 30)
}

// KAMABy Kaufman Adaptive Moving Average 并行计算版本
// period: 10, fast: 2, slow: 30
func KAMABy(data []float64, period int, fast, slow int) []float64 {
	n := len(data)
	res := make([]float64, n)

	if n < period+1 {
		// 数据长度不足，全部填充NaN
		for i := range res {
			res[i] = math.NaN()
		}
		return res
	}

	// 首先计算ER效率比率
	erValues := ER(data, period)

	// 预计算常量
	fastV := 2.0 / float64(fast+1)
	slowV := 2.0 / float64(slow+1)

	// 从第period+1个位置开始计算KAMA
	prevKAMA := math.NaN()
	for i := 0; i < n; i++ {
		effRatio := erValues[i]

		if math.IsNaN(effRatio) || math.IsNaN(data[i]) {
			res[i] = math.NaN()
			continue
		}

		// 计算alpha值
		alpha := math.Pow(effRatio*(fastV-slowV)+slowV, 2)

		// 如果前一个KAMA值是NaN，使用当前数据值作为初始值
		if math.IsNaN(prevKAMA) {
			prevKAMA = data[i-1]
		}
		// 应用KAMA公式: alpha * curVal + (1-alpha) * prevKAMA
		res[i] = alpha*data[i] + (1-alpha)*prevKAMA
		prevKAMA = res[i]
	}

	return res
}

// RMA 相对移动平均线，对应 stateful 指标实现中的 RMA
func RMA(data []float64, period int) []float64 {
	return RMABy(data, period, 0, math.NaN())
}

// RMABy 相对移动平均线，可指定初始化方式和初始值
func RMABy(data []float64, period int, initType int, initVal float64) []float64 {
	alpha := 1.0 / float64(period)
	return ewma(data, period, alpha, initType, initVal)
}

// SMA 简单移动平均，对应 stateful 指标实现中的 SMA
func SMA(data []float64, period int) []float64 {
	sums := Sum(data, period)
	res := make([]float64, len(sums))

	for i, sum := range sums {
		if math.IsNaN(sum) {
			res[i] = math.NaN()
		} else {
			res[i] = sum / float64(period)
		}
	}

	return res
}

// SSF is Ehlers' two-pole Super Smoother Filter.
func SSF(data []float64, period int) []float64 {
	n := len(data)
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}
	if period < 1 {
		return out
	}
	a := math.Exp(-1.4142135623730951 * math.Pi / float64(period))
	b := 2 * a * math.Cos(1.4142135623730951*math.Pi/float64(period))
	c2, c3 := b, -a*a
	c1 := 1 - c2 - c3
	for i := period - 1; i < n; i++ {
		if i == period-1 {
			out[i] = data[i]
		} else {
			p := out[i-1]
			pp := out[i-2]
			if math.IsNaN(pp) {
				pp = p
			}
			out[i] = c1*(data[i]+data[i-1])/2 + c2*p + c3*pp
		}
	}
	return out
}

// Sum 计算滑动窗口的累加和，遇到nan时跳过
func Sum(data []float64, period int) []float64 {
	n := len(data)
	res := make([]float64, n)

	var sum float64
	tmp := make([]float64, 0, period)

	for i, v := range data {
		// 如果当前输入值为 NaN，则跳过
		if math.IsNaN(v) {
			res[i] = math.NaN()
			continue
		}

		// 累加当前值
		sum += v
		tmp = append(tmp, v)

		// 如果连续有效的数据点数量超过窗口期，
		// 则减去滑出窗口的第一个元素的值。
		if len(tmp) > period {
			sum -= tmp[0]
			tmp = tmp[1:]
		}

		// 只有当连续有效的数据点数量达到窗口期时，才记录结果
		if len(tmp) >= period {
			res[i] = sum
		} else {
			res[i] = math.NaN()
		}
	}

	return res
}

// VIDYA is a variable-index dynamic average using CMO as the volatility index.
func VIDYA(data []float64, period int) []float64 {
	n := len(data)
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}
	if period < 1 {
		return out
	}
	alpha := 2.0 / float64(period+1)
	for i := period - 1; i < n; i++ {
		if i == period-1 {
			s := 0.
			for j := i - period + 1; j <= i; j++ {
				s += data[j]
			}
			out[i] = s / float64(period)
			continue
		}
		gain, loss := 0., 0.
		for j := i - period + 1; j <= i; j++ {
			d := data[j] - data[j-1]
			if d > 0 {
				gain += d
			} else {
				loss -= d
			}
		}
		cmo := 0.
		if gain+loss != 0 {
			cmo = math.Abs((gain - loss) / (gain + loss))
		}
		out[i] = out[i-1] + alpha*cmo*(data[i]-out[i-1])
	}
	return out
}

// VWMA 成交量加权移动平均线
// 这个版本会在遇到 NaN 时重置计算状态。
func VWMA(price []float64, volume []float64, period int) []float64 {
	n := len(price)
	res := make([]float64, n)

	// 状态变量：用于跟踪当前连续窗口
	sumCost := 0.0
	sumWeight := 0.0
	costs := make([]float64, 0, period)
	volumes := make([]float64, 0, period)

	// 步骤 3: 遍历所有数据点
	for i := 0; i < n; i++ {
		// 如果数据点有效，则累加
		cost := price[i] * volume[i]
		if math.IsNaN(cost) {
			res[i] = math.NaN()
			continue
		}
		sumCost += cost
		sumWeight += volume[i]
		costs = append(costs, cost)
		volumes = append(volumes, volume[i])

		if len(volumes) > period {
			sumCost -= costs[0]
			sumWeight -= volumes[0]
			costs = costs[1:]
			volumes = volumes[1:]
		}
		if len(volumes) >= period {
			res[i] = sumCost / sumWeight
		} else {
			res[i] = math.NaN()
		}
	}

	return res
}

// WMA an implementation of Weighted Moving Average.
// This version handles NaN values by resetting the calculation state.
// It remains high-performance for continuous non-NaN data segments.
func WMA(data []float64, period int) []float64 {
	n := len(data)
	res := make([]float64, n)

	// Handle cases with insufficient data or invalid period
	if n == 0 || period <= 0 {
		return res // Return empty or all-zero slice
	}

	// The sum of weights is a constant
	sumWei := float64(period) * float64(period+1) * 0.5

	// State variables for calculation
	var weightedSum float64
	var windowSum float64
	arr := make([]float64, 0, n)
	for i, val := range data {
		// Check for NaN input value
		if math.IsNaN(val) {
			res[i] = math.NaN()
			continue
		}
		arr = append(arr, val)
		if len(arr) > period {
			oldVal := arr[0]
			weightedSum -= windowSum
			windowSum -= oldVal
			arr = arr[1:]
		}
		arrNum := len(arr)
		weightedSum += val * float64(arrNum)
		windowSum += val

		if arrNum < period {
			res[i] = math.NaN()
		} else {
			res[i] = weightedSum / sumWei
		}
	}

	return res
}

func DEMA(data []float64, period int) []float64 {
	e := EMA(data, period)
	ee := EMA(e, period)
	r := make([]float64, len(data))
	for i := range r {
		r[i] = 2*e[i] - ee[i]
	}
	return r
}

func MAMA(data []float64, fast, slow float64) ([]float64, []float64) {
	p := int(math.Max(1, math.Round((fast+slow)/2)))
	m := EMA(data, p)
	f := EMA(m, p)
	return m, f
}

func SWMA(data []float64) []float64 {
	r := make([]float64, len(data))
	for i := range r {
		if i < 3 {
			r[i] = math.NaN()
		} else {
			r[i] = (data[i] + 2*data[i-1] + 2*data[i-2] + data[i-3]) / 6
		}
	}
	return r
}

func T3(data []float64, period int) []float64 {
	v := .7
	e1 := EMA(data, period)
	e2 := EMA(e1, period)
	e3 := EMA(e2, period)
	e4 := EMA(e3, period)
	e5 := EMA(e4, period)
	e6 := EMA(e5, period)
	c1 := -v * v * v
	c2 := 3*v*v + 3*v*v*v
	c3 := -6*v*v - 3*v - 3*v*v*v
	c4 := 1 + 3*v + v*v*v + 3*v*v
	r := make([]float64, len(data))
	for i := range r {
		r[i] = c1*e6[i] + c2*e5[i] + c3*e4[i] + c4*e3[i]
	}
	return r
}

func TRIMA(data []float64, period int) []float64 { return SMA(SMA(data, (period+1)/2), period/2+1) }

func ZLMA(data []float64, period int) []float64 {
	lag := (period - 1) / 2
	adj := make([]float64, len(data))
	for i := range data {
		if i < lag {
			adj[i] = math.NaN()
		} else {
			adj[i] = 2*data[i] - data[i-lag]
		}
	}
	return EMA(adj, period)
}

func ewma(data []float64, period int, alpha float64, initType int, initVal float64) []float64 {
	n := len(data)
	result := make([]float64, n)
	var prevVal = math.NaN()
	for i := 0; i < n; i++ {
		inVal := data[i]
		if math.IsNaN(inVal) {
			result[i] = inVal
			continue
		}
		curVal := math.NaN()
		if math.IsNaN(prevVal) {
			// 计算第一个有效 EMA 值
			if !math.IsNaN(initVal) {
				curVal = alpha*inVal + (1-alpha)*initVal
			} else if initType == 0 {
				curVal = computeFirstSMA(data, period, i)
			} else {
				curVal = inVal
			}
		} else {
			curVal = alpha*inVal + (1-alpha)*prevVal
		}
		result[i] = curVal
		if !math.IsNaN(curVal) {
			prevVal = curVal
		}
	}
	return result
}

// ER Efficiency Ratio / Trend to Noise Ratio 并行计算版本
// suggest period: 8
func ER(data []float64, period int) []float64 {
	n := len(data)
	res := make([]float64, n)

	sumVal := 0.0
	arr := make([]float64, 0, period)
	arrIn := make([]float64, 0, period+1)
	prevIn := math.NaN()
	for i := 0; i < n; i++ {
		inVal := data[i]
		chgVal := math.Abs(inVal - prevIn)
		if !math.IsNaN(inVal) {
			prevIn = inVal
			arrIn = append(arrIn, inVal)
		}
		if math.IsNaN(chgVal) {
			res[i] = math.NaN()
			continue
		}
		arr = append(arr, chgVal)
		sumVal += chgVal

		if len(arr) > period {
			sumVal -= arr[0]
			arr = arr[1:]
		}

		if len(arrIn) > period {
			periodVal := arrIn[0]
			arrIn = arrIn[1:]
			if sumVal > 0 {
				diffVal := math.Abs(inVal - periodVal)
				res[i] = diffVal / sumVal
				continue
			}
		}
		res[i] = math.NaN()
	}

	return res
}
