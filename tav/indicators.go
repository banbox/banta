package tav

import (
	"fmt"
	"math"
)

func HL2(a, b []float64) []float64 {
	res := make([]float64, len(a))
	for i, va := range a {
		res[i] = va*0.5 + b[i]*0.5
	}
	return res
}

func HLC3(a, b, c []float64) []float64 {
	res := make([]float64, len(a))
	for i, va := range a {
		res[i] = (va + b[i] + c[i]) / 3
	}
	return res
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

// SMA 简单移动平均，对应 sta_inds.go 中的 SMA
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

func computeFirstSMA(data []float64, period int, index int) float64 {
	if index+1 < period {
		return math.NaN()
	}
	sum := 0.0
	count := 0
	for i := max(0, index-period); i <= index; i++ {
		if !math.IsNaN(data[i]) {
			sum += data[i]
			count += 1
		}
	}
	if count < period {
		return math.NaN()
	}
	return sum / float64(count)
}

// EMA 指数移动平均线，对应 sta_inds.go 中的 EMA
func EMA(data []float64, period int) []float64 {
	return EMABy(data, period, 0)
}

// EMABy 指数移动平均线，可指定初始化方式
func EMABy(data []float64, period int, initType int) []float64 {
	alpha := 2.0 / float64(period+1)
	return ewma(data, period, alpha, initType, math.NaN())
}

// RMA 相对移动平均线，对应 sta_inds.go 中的 RMA
func RMA(data []float64, period int) []float64 {
	return RMABy(data, period, 0, math.NaN())
}

// RMABy 相对移动平均线，可指定初始化方式和初始值
func RMABy(data []float64, period int, initType int, initVal float64) []float64 {
	alpha := 1.0 / float64(period)
	return ewma(data, period, alpha, initType, initVal)
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

// TR 真实波动范围，对应 sta_inds.go 中的 TR
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

// ATR 平均真实波动范围，对应 sta_inds.go 中的 ATR
func ATR(high, low, close []float64, period int) []float64 {
	tr := TR(high, low, close)
	return RMA(tr, period)
}

// RSI 相对强度指数，对应 sta_inds.go 中的 RSI
func RSI(data []float64, period int) []float64 {
	return RSIBy(data, period, 0)
}

// RSIBy 带偏移量的RSI计算。
// 同样重构为分块处理逻辑，确保与带状态版本行为一致，并提升代码清晰度。
func RSIBy(data []float64, period int, subVal float64) []float64 {
	n := len(data)
	res := make([]float64, n)
	p := float64(period)

	var avgGain, avgLoss float64
	validCount := 0 // 连续有效差值的计数器

	// 从索引1开始，因为RSI基于价格变化
	if n > 0 {
		res[0] = math.NaN()
	}
	var prevSrc = math.NaN()
	for i := 0; i < n; i++ {
		// 如果当前或前一个值是NaN，则跳过
		if math.IsNaN(data[i]) || math.IsNaN(prevSrc) {
			if !math.IsNaN(data[i]) {
				prevSrc = data[i]
			}
			res[i] = math.NaN()
			continue
		}
		validCount++
		delta := data[i] - prevSrc
		prevSrc = data[i]
		var gainDelta, lossDelta float64
		if delta >= 0 {
			gainDelta = delta
		} else {
			lossDelta = -delta
		}

		// Wilder's Smoothing (EMA)
		if validCount > period {
			avgGain = (avgGain*float64(period-1) + gainDelta) / p
			avgLoss = (avgLoss*float64(period-1) + lossDelta) / p
		} else { // 初始SMA计算阶段
			avgGain += gainDelta
			avgLoss += lossDelta
			if validCount == period {
				avgGain /= p
				avgLoss /= p
			}
		}

		// 当有足够的周期数据后开始计算RSI值
		if validCount >= period {
			if avgGain+avgLoss == 0 {
				res[i] = 100 - subVal // 避免除以零，通常意味着无波动或纯粹上涨
			} else {
				res[i] = 100*avgGain/(avgGain+avgLoss) - subVal
			}
		} else {
			res[i] = math.NaN()
		}
	}
	return res
}

// MACD 移动平均线收敛发散指标，对应 sta_inds.go 中的 MACD
func MACD(data []float64, fast, slow, smooth int) ([]float64, []float64) {
	return MACDBy(data, fast, slow, smooth, 0)
}

// MACDBy 可自定义初始化方式的MACD
func MACDBy(data []float64, fast, slow, smooth, initType int) ([]float64, []float64) {
	n := len(data)
	macd := make([]float64, n)
	signal := make([]float64, n)

	// 计算快慢均线
	fastEma := EMABy(data, fast, initType)
	slowEma := EMABy(data, slow, initType)

	// 计算MACD线
	for i := 0; i < n; i++ {
		if !math.IsNaN(fastEma[i]) && !math.IsNaN(slowEma[i]) {
			macd[i] = fastEma[i] - slowEma[i]
		} else {
			macd[i] = math.NaN()
		}
	}

	// 计算信号线
	signal = EMABy(macd, smooth, initType)

	return macd, signal
}

// StdDev 标准差，对应 sta_inds.go 中的 StdDev
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

// BBANDS 布林带指标，对应 sta_inds.go 中的 BBANDS
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

// TD Tom DeMark Sequence 狄马克序列，对应sta_inds.go中的TD
func TD(data []float64) []float64 {
	n := len(data)
	res := make([]float64, n)

	winSize := 5
	prevs := make([]float64, 0, winSize)
	for i := 0; i < n; i++ {
		inVal := data[i]
		if math.IsNaN(inVal) {
			res[i] = math.NaN()
			continue
		}
		prevs = append(prevs, inVal)
		if len(prevs) > winSize {
			prevs = prevs[1:]
		}
		if len(prevs) < winSize {
			res[i] = math.NaN()
			continue
		}

		sub4 := inVal - prevs[0]
		step := 1.0
		if sub4 == 0 {
			step = 0
		} else if sub4 < 0 {
			step = -1
		}

		// 获取前一个计算出的有效TD值
		prevNum := res[i-1]

		// 如果前一个值有效且与当前趋势同向，则累加
		if !math.IsNaN(prevNum) && prevNum*step > 0 {
			res[i] = math.Round(prevNum) + step
		} else {
			// 否则，重置计数器为step(1或-1)
			res[i] = step
		}
	}

	return res
}

// ADX 平均方向指数，对应 sta_inds.go 中的 ADX
func ADX(high, low, close []float64, period int) []float64 {
	return ADXBy(high, low, close, period, 0, 0)
}

// ADXBy 可自定义方法的平均方向指数
// method=0 经典ADX计算方法
// method=1 TradingView "ADX and DI for v4"方法
// smoothing=0 表示使用period作为平滑周期
// 返回 adx值数组
func ADXBy(high, low, close []float64, period, smoothing, method int) []float64 {
	n := len(high)
	adx := make([]float64, n)

	if n < period+1 {
		for i := range adx {
			adx[i] = math.NaN()
		}
		return adx
	}

	// 计算+DI和-DI
	plusDI, minusDI := pluMinDIBy(high, low, close, period, method)

	// 设置平滑周期，如果未指定则使用period
	if smoothing == 0 {
		smoothing = period
	}

	// 计算DX: abs(+DI - -DI)/(+DI + -DI) * 100
	dx := make([]float64, n)
	for i := 0; i < n; i++ {
		if !math.IsNaN(plusDI[i]) && !math.IsNaN(minusDI[i]) && (plusDI[i]+minusDI[i]) > 0 {
			dx[i] = math.Abs(plusDI[i]-minusDI[i]) / (plusDI[i] + minusDI[i]) * 100
		} else {
			dx[i] = math.NaN()
		}
	}

	// 根据method选择平滑方式计算ADX
	if method == 0 {
		// 使用RMA平滑
		adx = RMA(dx, smoothing)
	} else {
		// 使用SMA平滑
		adx = SMA(dx, smoothing)
	}

	return adx
}

func PluMinDI(high, low, close []float64, period int) ([]float64, []float64) {
	return pluMinDIBy(high, low, close, period, 0)
}

// pluMinDIBy 计算方向指标+DI和-DI
func pluMinDIBy(high, low, close []float64, period, method int) ([]float64, []float64) {
	n := len(high)
	plusDI := make([]float64, n)
	minusDI := make([]float64, n)

	// 计算+DM和-DM
	plusDM, minusDM, trDM := pluMinDMBy(high, low, close, period, method)

	// 计算+DI和-DI
	for i := 0; i < n; i++ {
		trVal := trDM[i]
		if i >= period && !math.IsNaN(trVal) && trVal > 0 {
			plusDI[i] = 100 * plusDM[i] / trVal
			minusDI[i] = 100 * minusDM[i] / trVal
		} else {
			plusDI[i] = math.NaN()
			minusDI[i] = math.NaN()
		}
	}

	return plusDI, minusDI
}

func PluMinDM(high, low, cls []float64, period int) ([]float64, []float64) {
	plus, minu, _ := pluMinDMBy(high, low, cls, period, 0)
	return plus, minu
}

// pluMinDMBy 计算正负方向移动 (+DM, -DM) 的并行版本。
// method=0: 经典方式, 使用 period 作为初始化长度。
// method=1: TradingView方式, 使用 period+1 作为初始化长度。
// 返回平滑后的+DM和平滑后的-DM两个切片。
func pluMinDMBy(high, low, close []float64, period, method int) ([]float64, []float64, []float64) {
	n := len(close)
	// 初始化结果切片
	resPlusDM := make([]float64, n)
	resMinusDM := make([]float64, n)
	resTRDM := make([]float64, n)

	// 根据方法确定初始化长度
	initLen := period
	if method == 1 {
		initLen = period + 1
	}

	var currentPlusMA, currentMinusMA, trMA float64
	alpha := 1.0 / float64(period)

	// 在单次循环中计算并平滑指标
	num := 0
	c1 := math.NaN()
	for i := 0; i < n; i++ {
		var dmhVal, dmlVal, tr = math.NaN(), math.NaN(), math.NaN()
		if i > 0 {
			dmhVal = high[i] - high[i-1]
			dmlVal = low[i-1] - low[i]
			h, l := high[i], low[i]
			tr = max(h-l, math.Abs(h-c1), math.Abs(l-c1))
		}
		c0 := close[i]
		if math.IsNaN(c0) || math.IsNaN(tr) {
			if !math.IsNaN(c0) {
				c1 = c0
			}
			resPlusDM[i] = math.NaN()
			resMinusDM[i] = math.NaN()
			resTRDM[i] = math.NaN()
			continue
		}
		c1 = c0
		num += 1

		plusDM, minusDM := 0.0, 0.0
		if dmhVal > max(dmlVal, 0) {
			plusDM = dmhVal
		} else if dmlVal > max(dmhVal, 0) {
			minusDM = dmlVal
		}

		if num <= initLen-1 {
			currentPlusMA += plusDM
			currentMinusMA += minusDM
			trMA += tr
			if num <= period-1 {
				resPlusDM[i] = math.NaN()
				resMinusDM[i] = math.NaN()
				resTRDM[i] = math.NaN()
			} else {
				resPlusDM[i] = currentPlusMA
				resMinusDM[i] = currentMinusMA
				resTRDM[i] = trMA
			}
		} else {
			// 应用Wilder平滑公式
			currentPlusMA = currentPlusMA*(1-alpha) + plusDM
			currentMinusMA = currentMinusMA*(1-alpha) + minusDM
			trMA = trMA*(1-alpha) + tr
			resPlusDM[i] = currentPlusMA
			resMinusDM[i] = currentMinusMA
			resTRDM[i] = trMA
		}
	}

	return resPlusDM, resMinusDM, resTRDM
}

// slidingWindow is a generic helper for Highest/Lowest using a deque.
// This function already handles NaNs correctly and does not need changes.
func slidingWindow(data []float64, period int, findMax bool) []float64 {
	n := len(data)
	if period <= 0 {
		// For invalid period, return NaNs for all points.
		res := make([]float64, n)
		for i := range res {
			res[i] = math.NaN()
		}
		return res
	}

	result := make([]float64, n)
	deque := make([]int, 0, period) // Deque stores indices of elements.

	// Define comparison based on whether we are finding max or min.
	var compare func(a, b float64) bool
	if findMax {
		compare = func(a, b float64) bool { return a >= b }
	} else {
		compare = func(a, b float64) bool { return a <= b }
	}

	validNum := 0
	winStart := -period
	flags := make([]int, n) // 0表示正常值，1表示nan；遇到nan记录，左侧边界到达时+1
	for i := 0; i < n; i++ {
		val := data[i]
		if math.IsNaN(val) {
			result[i] = math.NaN()
			flags[i] = 1 //winStart到此位置时应往前+1
			continue
		} else if winStart >= 0 {
			f := flags[winStart]
			for f > 0 {
				winStart += f
				f = flags[winStart]
			}
		}
		winStart += 1
		// Remove indices from the front that are outside the current window.
		if len(deque) > 0 && deque[0] < winStart {
			deque = deque[1:]
		}
		// Maintain monotonic property of the deque.
		// Note: If data[deque[len(deque)-1]] were NaN, it wouldn't be in the deque.
		for len(deque) > 0 && compare(val, data[deque[len(deque)-1]]) {
			deque = deque[:len(deque)-1]
		}
		deque = append(deque, i)
		validNum += 1
		// A result can be computed only when the window is full.
		if validNum >= period && len(deque) > 0 {
			result[i] = data[deque[0]]
		} else {
			result[i] = math.NaN()
		}
	}
	return result
}

// Highest finds the highest value over a preceding period for each point.
// No changes needed as it relies on the correct slidingWindow implementation.
func Highest(data []float64, period int) []float64 {
	return slidingWindow(data, period, true)
}

// Lowest finds the lowest value over a preceding period for each point.
// No changes needed as it relies on the correct slidingWindow implementation.
func Lowest(data []float64, period int) []float64 {
	return slidingWindow(data, period, false)
}

// findExtremeBarOffset finds the offset of the highest/lowest bar in a preceding period.
// This function has been rewritten to correctly handle NaN values.
func findExtremeBarOffset(input []float64, period int, findLowest bool) []float64 {
	n := len(input)
	output := make([]float64, n)

	// If period is invalid or 0, or no data, all results are NaN.
	if period <= 0 || n == 0 {
		for i := 0; i < n; i++ {
			output[i] = math.NaN()
		}
		return output
	}

	for i := 0; i < n; i++ {
		// Not enough preceding data points for a full window ending at i.
		if i < period-1 || math.IsNaN(input[i]) {
			output[i] = math.NaN()
			continue
		}

		// Initialize with sentinel values, similar to the stateful version.
		extremeVal := math.NaN()
		extremeOffset := -1

		// Iterate through the window of data: input[i-period+1 ... i]
		// k represents the "bars ago" offset, from 0 (current) to period-1.
		checkNum := 0
		for k := 0; i-k >= 0 && checkNum < period; k++ {
			valInWindow := input[i-k] // Value at offset k (k bars ago)

			// If current value is NaN, skip it, just like the stateful version.
			if math.IsNaN(valInWindow) {
				continue
			}
			checkNum += 1

			if findLowest { // For LowestBar
				// If this is the first non-NaN value found (extremeOffset == -1)
				// or if the current value is lower than the recorded minimum.
				if extremeOffset == -1 || valInWindow < extremeVal {
					extremeVal = valInWindow
					extremeOffset = k
				}
			} else { // For HighestBar
				// If this is the first non-NaN value found (extremeOffset == -1)
				// or if the current value is higher than the recorded maximum.
				if extremeOffset == -1 || valInWindow > extremeVal {
					extremeVal = valInWindow
					extremeOffset = k
				}
			}
		}

		// After checking the whole window, if extremeOffset is still -1,
		// it means all values in the window were NaN.
		if extremeOffset == -1 || checkNum < period {
			output[i] = math.NaN()
		} else {
			output[i] = float64(extremeOffset)
		}
	}
	return output
}

// HighestBar finds the offset of the highest bar in a preceding period.
func HighestBar(data []float64, period int) []float64 {
	return findExtremeBarOffset(data, period, false)
}

// LowestBar finds the offset of the lowest bar in a preceding period.
func LowestBar(data []float64, period int) []float64 {
	return findExtremeBarOffset(data, period, true)
}

// Stoch 计算随机指标(K值)。
// 优化了NaN处理和分支逻辑，使其更清晰且不易出错。
func Stoch(high, low, close []float64, period int) []float64 {
	n := len(close)
	if n < period || period <= 0 {
		res := make([]float64, n)
		for i := range res {
			res[i] = math.NaN()
		}
		return res
	}

	res := make([]float64, n)
	minLen := min(len(high), len(low), n)

	hh := Highest(high, period)
	ll := Lowest(low, period)

	for i := 0; i < minLen; i++ {
		// 如果当前收盘价无效，则结果也无效
		maxChg := hh[i] - ll[i]
		if math.IsNaN(close[i]) || math.IsNaN(maxChg) {
			res[i] = math.NaN()
			continue
		}

		if equalNearly(maxChg, 0) {
			res[i] = 50.0
		} else {
			res[i] = (close[i] - ll[i]) / maxChg * 100
		}
	}
	// 填充剩余部分
	for i := minLen; i < n; i++ {
		res[i] = math.NaN()
	}

	return res
}

// KDJ 随机指标KDJ，对应 sta_inds.go 中的 KDJ
func KDJ(high, low, close []float64, period, sm1, sm2 int) ([]float64, []float64, []float64) {
	return KDJBy(high, low, close, period, sm1, sm2, "rma")
}

func KDJBy(high, low, close []float64, period int, sm1 int, sm2 int, maBy string) (k, d, rsv []float64) {
	n := len(close)
	// Basic validation
	if n == 0 || period <= 0 || sm1 <= 0 || sm2 <= 0 {
		k = make([]float64, n)
		d = make([]float64, n)
		rsv = make([]float64, n)
		for i := 0; i < n; i++ {
			k[i], d[i], rsv[i] = math.NaN(), math.NaN(), math.NaN()
		}
		return
	}

	// Ensure consistent lengths for input slices
	minLen := min(len(high), len(low), n)
	if minLen != n {
		n = minLen
		high = high[:n]
		low = low[:n]
		close = close[:n]
	}

	// Initialize result slices
	k = make([]float64, n)
	d = make([]float64, n)

	// 1. Calculate RSV
	rsv = Stoch(high, low, close, period)

	// 2. Calculate K, d
	switch maBy {
	case "rma":
		k = RMABy(rsv, sm1, 0, 50.0)
		d = RMABy(k, sm2, 0, 50.0)
	case "sma":
		k = SMA(rsv, sm1)
		d = SMA(k, sm2)
	default:
		panic(fmt.Sprintf("unknown maBy for KDJByParallel: %s", maBy))
	}
	return k, d, rsv
}

// Aroon 阿隆指标，对应 sta_inds.go 中的 Aroon
func Aroon(high, low []float64, period int) ([]float64, []float64, []float64) {
	n := len(high)
	up := make([]float64, n)
	down := make([]float64, n)
	osc := make([]float64, n)

	if n < period+1 {
		for i := range up {
			up[i] = math.NaN()
			down[i] = math.NaN()
			osc[i] = math.NaN()
		}
		return up, osc, down
	}

	// 计算高点和低点的位置
	highestBars := HighestBar(high, period+1)
	lowestBars := LowestBar(low, period+1)

	// 计算Aroon指标
	factor := -100 / float64(period)
	for i := period; i < n; i++ {
		if !math.IsNaN(highestBars[i]) && !math.IsNaN(lowestBars[i]) {
			up[i] = factor*highestBars[i] + 100
			down[i] = factor*lowestBars[i] + 100
			osc[i] = up[i] - down[i]
		} else {
			up[i] = math.NaN()
			down[i] = math.NaN()
			osc[i] = math.NaN()
		}
	}

	// 填充前面的NaN
	for i := 0; i < period; i++ {
		up[i] = math.NaN()
		down[i] = math.NaN()
		osc[i] = math.NaN()
	}

	return up, osc, down
}

// ROC 变化率指标，对应 sta_inds.go 中的 ROC
func ROC(data []float64, period int) []float64 {
	n := len(data)
	res := make([]float64, n)

	prevs := make([]float64, 0, n)
	// 计算ROC值
	for i := 0; i < n; i++ {
		val := data[i]
		if math.IsNaN(val) {
			res[i] = math.NaN()
			continue
		}
		prevs = append(prevs, val)
		if len(prevs) > period+1 {
			prevs = prevs[1:]
		} else if len(prevs) <= period {
			res[i] = math.NaN()
			continue
		}
		preVal := prevs[0]

		if preVal != 0 {
			res[i] = (val - preVal) / preVal * 100
		} else {
			res[i] = math.NaN() // 避免除以0
		}
	}

	return res
}

// UpDown 上下变动指标，对应 sta_inds.go 中的 UpDown
// 修正版：正确处理NaN值，并在遇到NaN时重置累计状态。
func UpDown(data []float64, vtype int) []float64 {
	n := len(data)
	res := make([]float64, n)

	for i := range res {
		res[i] = math.NaN()
	}

	if n < 2 {
		return res
	}

	var prev float64 = 0.0
	for i := 1; i < n; i++ {
		if math.IsNaN(data[i]) || math.IsNaN(data[i-1]) {
			prev = 0.0
			continue
		}

		if math.IsNaN(res[i-1]) {
			// 根据定义，序列的第一个基准点的UpDown值应为0。
			res[i-1] = 0.0
		}

		// --- 以下是原始的核心计算逻辑，无需改动 ---
		sub := data[i] - data[i-1]
		if sub == 0 {
			prev = 0
		} else if sub > 0 {
			if prev > 0 && vtype == 0 {
				prev += 1
			} else {
				prev = 1
			}
		} else { // sub < 0
			if prev < 0 && vtype == 0 {
				prev -= 1
			} else {
				prev = -1
			}
		}
		// 将计算结果存入res数组
		res[i] = prev
	}

	return res
}

func PercentRank(data []float64, period int) []float64 {
	n := len(data)
	res := make([]float64, n)

	vals := make([]float64, 0, period)
	// 计算每个滑动窗口的 PercentRank，从位置 period-1 开始
	for i := 0; i < n; i++ {
		curV := data[i]
		if math.IsNaN(curV) {
			res[i] = math.NaN()
			continue
		}
		vals = append(vals, curV)
		if len(vals) > period {
			vals = vals[1:]
		} else if len(vals) < period {
			res[i] = math.NaN()
			continue
		}

		lowNum := float64(0)
		// Iterate over the values in the current window to count how many are less than or equal to curV
		for j := 0; j < period-1; j++ {
			if vals[j] <= curV {
				lowNum += 1
			}
		}

		// Calculate the percentile rank and store it in the result slice
		res[i] = lowNum * 100 / float64(period)
	}
	return res
}

// CRSI Connors RSI，对应 sta_inds.go 中的 CRSI
func CRSI(data []float64, period, upDn, rocVal int) []float64 {
	return CRSIBy(data, period, upDn, rocVal, 0)
}

// CRSIBy 可自定义计算方法的Connors RSI
func CRSIBy(data []float64, period, upDn, rocVal, vtype int) []float64 {
	n := len(data)
	res := make([]float64, n)
	if n == 0 {
		return res
	}
	// 计算RSI
	rsi := RSI(data, period)

	// 计算UpDown的RSI
	updown := UpDown(data, vtype)
	ud := RSI(updown, upDn)

	// 计算ROC或PercentRank
	var rc []float64
	if vtype == 0 {
		// TradingView方法：使用PercentRank(ROC(data, 1), roc)
		rc = PercentRank(ROC(data, 1), rocVal)
	} else {
		// ta-lib社区方法：直接使用ROC(data, roc)
		rc = ROC(data, rocVal)
	}

	// 计算CRSI
	for i := 0; i < n; i++ {
		if !math.IsNaN(rsi[i]) && !math.IsNaN(ud[i]) && !math.IsNaN(rc[i]) {
			res[i] = (rsi[i] + ud[i] + rc[i]) / 3
		} else {
			res[i] = math.NaN()
		}
	}

	return res
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

// AvgDev 平均偏差，对应 sta_inds.go 中的 AvgDev
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

// CCI 商品通道指数，对应 sta_inds.go 中的 CCI
func CCI(data []float64, period int) []float64 {
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
	avgDev := AvgDev(data, period)

	// 计算CCI
	for i := period - 1; i < n; i++ {
		if !math.IsNaN(sma[i]) && !math.IsNaN(avgDev[i]) && avgDev[i] != 0 {
			res[i] = (data[i] - sma[i]) / (0.015 * avgDev[i])
		} else {
			res[i] = math.NaN()
		}
	}

	// 填充前面的NaN
	for i := 0; i < period-1; i++ {
		res[i] = math.NaN()
	}

	return res
}

func moveToFront(data []float64, i int) []float64 {
	res := make([]float64, 0, len(data))
	res = append(res, data[i])
	res = append(res, data[:i]...)
	if i+1 < len(data) {
		res = append(res, data[i+1:]...)
	}
	return res
}

func moveToEnd(data []float64, i int) []float64 {
	res := make([]float64, 0, len(data))
	res = append(res, data[:i]...)
	if i+1 < len(data) {
		res = append(res, data[i+1:]...)
	}
	res = append(res, data[i])
	return res
}

// CMF Chaikin Money Flow
// period: 20
func CMF(high, low, close, volume []float64, period int) []float64 {
	n := len(close)
	res := make([]float64, n)

	// Not enough data to calculate
	if n < period {
		for i := range res {
			res[i] = math.NaN()
		}
		return res
	}

	// Pass 1: Calculate Money Flow Volume for each bar
	mfv := make([]float64, n)
	for i := 0; i < n; i++ {
		h, l, c := high[i], low[i], close[i]
		hilo := h - l

		if hilo > 0 {
			// Money Flow Multiplier = [(Close - Low) - (High - Close)] / (High - Low)
			multiplier := ((c - l) - (h - c)) / hilo
			// Money Flow Volume = Money Flow Multiplier x Volume
			mfv[i] = multiplier * volume[i]
		} else {
			mfv[i] = 0.0 // MFV is 0 if high equals low
		}
	}

	var mfvSum, volSum float64

	// Pass 2: Use a sliding window to calculate CMF
	okNum := 0
	for i := 0; i < n; i++ {
		mfvVal, vol := mfv[i], volume[i]
		if math.IsNaN(mfvVal) || math.IsNaN(vol) || vol <= 0 {
			res[i] = math.NaN()
			mfv = moveToFront(mfv, i)
			volume = moveToFront(volume, i)
			continue
		}
		mfvSum += mfvVal
		volSum += vol
		okNum += 1

		// Subtract the value that fell out of the window
		if okNum > period {
			mfvSum -= mfv[i-period]
			volSum -= volume[i-period]
		}

		if okNum >= period {
			// CMF = Sum(Money Flow Volume) / Sum(Volume)
			res[i] = mfvSum / volSum
		} else {
			// Fill initial values with NaN
			res[i] = math.NaN()
		}
	}
	return res
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

// Williams %R 并行计算版本
// suggest period: 14
func WillR(high, low, close []float64, period int) []float64 {
	n := len(close)
	if n == 0 || period <= 0 {
		return make([]float64, 0)
	}

	// 确保所有数组长度一致
	minLen := min(len(high), len(low), n)
	if minLen < n {
		n = minLen
		high, low, close = high[:n], low[:n], close[:n]
	}

	res := make([]float64, n)
	hh, ll := Highest(high, period), Lowest(low, period)

	for i := 0; i < n; i++ {
		// 如果当前收盘价无效，则结果也无效
		if math.IsNaN(close[i]) {
			res[i] = math.NaN()
			continue
		}

		rangeHL := hh[i] - ll[i]
		// 如果最高最低价无效，或区间为0，则结果为NaN
		if math.IsNaN(rangeHL) || rangeHL == 0 {
			res[i] = math.NaN()
			continue
		}

		// 计算Williams %R
		res[i] = (close[i] - hh[i]) / rangeHL * 100
	}

	return res
}

// StochRSI StochasticRSI，对应 sta_inds.go 中的 StochRSI
// rsiLen: RSI周期, stochLen: Stoch周期, maK: K线SMA周期, maD: D线SMA周期
// 返回 [fastK, fastD]
func StochRSI(obj []float64, rsiLen int, stochLen int, maK int, maD int) ([]float64, []float64) {
	// 1. 计算RSI
	rsi := RSI(obj, rsiLen)

	// 2. 对RSI结果计算Stoch
	// Stoch的 high, low, close 输入都使用 rsi
	stochVal := Stoch(rsi, rsi, rsi, stochLen)

	// 3. 计算 fastK
	fastK := SMA(stochVal, maK)

	// 4. 计算 fastD
	fastD := SMA(fastK, maD)

	return fastK, fastD
}

// MFI 资金流量指数 - 并行计算版本
func MFI(high, low, close, volume []float64, period int) []float64 {
	n := len(close)
	res := make([]float64, n)

	// 计算正负资金流量
	posArr := make([]float64, 0, n)
	negArr := make([]float64, 0, n)

	// 计算滑动窗口的正负资金流量总和
	var sumPos, sumNeg float64

	prevPrice := math.NaN()
	for i := 0; i < n; i++ {
		cPrice := (high[i] + low[i] + close[i]) / 3.0
		if math.IsNaN(cPrice) {
			res[i] = math.NaN()
			continue
		}
		moneyFlow := cPrice * volume[i]
		var posFlow, negFlow float64
		if cPrice > prevPrice {
			posFlow = moneyFlow
		} else if cPrice < prevPrice {
			negFlow = moneyFlow
		}
		prevPrice = cPrice
		if math.IsNaN(moneyFlow) {
			res[i] = math.NaN()
			continue
		}
		posArr = append(posArr, posFlow)
		negArr = append(negArr, negFlow)
		sumPos += posFlow
		sumNeg += negFlow
		if len(posArr) >= period {
			if len(posArr) > period {
				// 滑动窗口：减去最旧的值
				sumPos -= posArr[0]
				sumNeg -= negArr[0]
				posArr = posArr[1:]
				negArr = negArr[1:]
			}
			// 计算MFI值
			if sumNeg > 0 {
				moneyFlowRatio := sumPos / sumNeg
				res[i] = 100 - (100 / (1 + moneyFlowRatio))
			} else {
				res[i] = math.NaN()
			}
		} else {
			res[i] = math.NaN()
		}
	}
	return res
}

// calculateRMIValues 根据up和down计算RMI值
func calculateRMIValues(up, down []float64) []float64 {
	n := len(up)
	rmi := make([]float64, n)

	for i := 0; i < n; i++ {
		if math.IsNaN(up[i]) || math.IsNaN(down[i]) {
			rmi[i] = math.NaN()
		} else if down[i] == 0 {
			rmi[i] = 100
		} else if up[i] == 0 {
			rmi[i] = 0
		} else {
			rmi[i] = 100 - (100 / (1 + up[i]/down[i]))
		}
	}

	return rmi
}

// RMI Relative Momentum Index 相对动量指标的并行计算版本
func RMI(data []float64, period int, montLen int) []float64 {
	n := len(data)
	if n == 0 {
		return make([]float64, 0)
	}

	// 计算价格变化，分离正负变化
	maxChg := make([]float64, n)
	minChg := make([]float64, n)

	// 计算变化值
	arr := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		inVal := data[i]
		if math.IsNaN(inVal) {
			maxChg[i] = math.NaN()
			minChg[i] = math.NaN()
			continue
		}
		arr = append(arr, inVal)
		if len(arr) < montLen+1 {
			maxChg[i] = math.NaN()
			minChg[i] = math.NaN()
			continue
		}
		if len(arr) > montLen+1 {
			arr = arr[1:]
		}
		chgVal := inVal - arr[0]
		maxChg[i] = math.Max(0, chgVal)
		minChg[i] = -math.Min(0, chgVal)
	}

	// 使用RMA计算平滑后的up和down
	up := RMABy(maxChg, period, 0, math.NaN())
	down := RMABy(minChg, period, 0, math.NaN())

	// 计算最终的RMI值
	return calculateRMIValues(up, down)
}

// precomputeLinRegConstants 预计算线性回归的常数部分
func precomputeLinRegConstants(period int) (sumX, sumX2, periodF float64) {
	periodF = float64(period)
	sumX = periodF * float64(period+1) * 0.5
	sumX2 = sumX * (2*periodF + 1) / 3
	return sumX, sumX2, periodF
}

// computeLinRegCoeff 计算线性回归系数
func computeLinRegCoeff(arr []float64, period int, sumX, sumX2, periodF float64) (slope, intercept, r float64) {
	var sumY, sumXY, sumY2 float64

	// 一次循环计算所有需要的和
	for i := 0; i < period; i++ {
		v := arr[i]
		sumY += v
		sumXY += float64(i+1) * v
		sumY2 += v * v
	}

	divisor := periodF*sumX2 - sumX*sumX
	slope = (periodF*sumXY - sumX*sumY) / divisor
	intercept = (sumY*sumX2 - sumX*sumXY) / divisor

	// 计算相关系数 r
	rn := periodF*sumXY - sumX*sumY
	rd := math.Sqrt(divisor * (periodF*sumY2 - sumY*sumY))
	r = rn / rd

	return slope, intercept, r
}

// LinRegAdv 计算线性回归的各种指标，并行版本
func LinReg(data []float64, period int) []float64 {
	return LinRegAdv(data, period, false, false, false, false, false, false)
}

// LinRegAdv 计算线性回归的各种指标，并行版本
func LinRegAdv(data []float64, period int, angle, intercept, degrees, r, slope, tsf bool) []float64 {
	n := len(data)
	result := make([]float64, n)

	// 预计算常数
	sumX, sumX2, periodF := precomputeLinRegConstants(period)

	// 滑动窗口计算
	arr := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		// 获取当前窗口的数据
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

		// 计算线性回归系数
		m, b, rVal := computeLinRegCoeff(arr, period, sumX, sumX2, periodF)

		// 根据需要的输出类型返回相应的值
		if slope {
			result[i] = m
		} else if intercept {
			result[i] = b
		} else if angle {
			theta := math.Atan(m)
			if degrees {
				theta *= 180 / math.Pi
			}
			result[i] = theta
		} else if r {
			result[i] = rVal
		} else if tsf {
			result[i] = m*periodF + b
		} else {
			result[i] = m*(periodF-1) + b
		}
	}

	return result
}

// CTI Correlation Trend Indicator 并行版本
// CTI是一个由John Ehler在2020年创建的振荡器
// 它根据价格范围内的价格接近正斜率或负斜率直线的程度来分配值
// 值范围从-1到1，建议周期：20
// 当前是性能优化版本，快20%；等同于LinRegAdv(data, period, false, false, false, true, false, false)
func CTI(data []float64, period int) []float64 {
	n := len(data)
	result := make([]float64, n)

	// 预计算常数（只计算CTI需要的部分）
	periodF := float64(period)
	sumX := periodF * float64(period+1) * 0.5
	sumX2 := sumX * (2*periodF + 1) / 3
	divisor := periodF*sumX2 - sumX*sumX

	arr := make([]float64, 0, n)
	// 滑动窗口计算，专门为相关系数r优化
	for i := 0; i < n; i++ {
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
		// 检查窗口中是否有NaN值，同时计算所需的和
		var sumY, sumXY, sumY2 float64

		for j, v := range arr {
			sumY += v
			sumXY += float64(j+1) * v
			sumY2 += v * v
		}

		// 计算相关系数 r
		rn := periodF*sumXY - sumX*sumY
		rd := math.Sqrt(divisor * (periodF*sumY2 - sumY*sumY))
		result[i] = rn / rd
	}

	return result
}

// wilderSmoothing 计算Wilder平滑，正确处理数据中出现的NaN。
func wilderSmoothing(data []float64, period int) []float64 {
	n := len(data)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}

	// 周期小于等于1无意义
	if n < period || period <= 1 {
		return result
	}

	alpha := 1.0 / float64(period)
	var sum float64 = 0.0
	var periodCount int = 0
	// lastSmoothed 用于保存上一个有效的平滑值
	var lastSmoothed float64 = math.NaN()

	for i, val := range data {
		if math.IsNaN(val) {
			// 如果在累加初始周期时遇到NaN，则重置累加
			if math.IsNaN(lastSmoothed) {
				sum = 0.0
				periodCount = 0
			}
			// 跳过当前NaN，lastSmoothed值被保留，用于下一个有效值的计算
			continue
		}

		// 检查是否已开始平滑计算
		if math.IsNaN(lastSmoothed) {
			// 累加初始周期的数据以计算第一个SMA
			sum += val
			periodCount++
			if periodCount == period {
				// 计算第一个SMA值，并将其作为后续平滑计算的起点
				lastSmoothed = sum / float64(period)
				result[i] = lastSmoothed
			}
		} else {
			// 使用Wilder平滑公式进行计算
			lastSmoothed = alpha*val + (1-alpha)*lastSmoothed
			result[i] = lastSmoothed
		}
	}
	return result
}

// CMO 计算Chande Momentum Oscillator (ta-lib版本)
func CMO(data []float64, period int) []float64 {
	return CMOBy(data, period, 0)
}

// CMOBy 计算Chande Momentum Oscillator
// maType: 0: ta-lib (Wilder's smoothing)   1: tradingView (simple moving sum)
func CMOBy(data []float64, period int, maType int) []float64 {
	n := len(data)
	result := make([]float64, n)

	if n < 2 || period <= 0 {
		for i := range result {
			result[i] = math.NaN()
		}
		return result
	}

	// 计算价格差值
	posVals, negVals := make([]float64, 0, n), make([]float64, 0, n)
	prevVal := data[0]
	for i := 1; i < n; i++ {
		inVal := data[i]
		diff := inVal - prevVal
		if !math.IsNaN(inVal) {
			prevVal = inVal
		}
		if math.IsNaN(diff) {
			posVals = append(posVals, math.NaN())
			negVals = append(negVals, math.NaN())
		} else if diff > 0 {
			posVals = append(posVals, diff)
			negVals = append(negVals, 0)
		} else {
			posVals = append(posVals, 0)
			negVals = append(negVals, -diff)
		}
	}
	var sumPos, sumNeg []float64

	if maType == 0 {
		// ta-lib: Wilder's smoothing
		sumPos = wilderSmoothing(posVals, period)
		sumNeg = wilderSmoothing(negVals, period)
	} else {
		// tradingView: simple moving sum
		sumPos = Sum(posVals, period)
		sumNeg = Sum(negVals, period)
	}

	// 计算CMO值
	result[0] = math.NaN() // 第一个值总是NaN，因为没有差值

	for i := 1; i < n; i++ {
		diffIdx := i - 1
		if diffIdx >= len(sumPos) || diffIdx >= len(sumNeg) {
			result[i] = math.NaN()
			continue
		}

		pos := sumPos[diffIdx]
		neg := sumNeg[diffIdx]

		if math.IsNaN(pos) || math.IsNaN(neg) || (pos+neg) == 0 {
			result[i] = math.NaN()
		} else {
			result[i] = (pos - neg) * 100 / (pos + neg)
		}
	}

	return result
}

// CHOP 乔普指数，对应 sta_inds.go 中的 CHOP
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

// calculateBounds 计算边界值：SMA - StdDev * 0.2
func calculateBounds(data []float64, maLen int) []float64 {
	n := len(data)
	bounds := make([]float64, n)

	if n < maLen {
		for i := range bounds {
			bounds[i] = math.NaN()
		}
		return bounds
	}

	// 同时计算SMA和StdDev以提高效率
	smaValues := SMA(data, maLen)
	stdDevValues := StdDev(data, maLen)

	for i := range bounds {
		if math.IsNaN(smaValues[i]) || math.IsNaN(stdDevValues[i]) {
			bounds[i] = math.NaN()
		} else {
			bounds[i] = smaValues[i] - stdDevValues[i]*0.2
		}
	}

	return bounds
}

// calculateRawGreaterThan 计算原始值大于边界的百分比值
func calculateRawGreaterThan(data, bounds []float64, stiffLen int) []float64 {
	n := len(data)
	rawGt := make([]float64, n)
	percentValue := 100.0 / float64(stiffLen)

	for i := 0; i < n; i++ {
		if math.IsNaN(data[i]) || math.IsNaN(bounds[i]) {
			rawGt[i] = math.NaN()
		} else if data[i] > bounds[i] {
			rawGt[i] = percentValue
		} else {
			rawGt[i] = 0.0
		}
	}

	return rawGt
}

/*
DV2 Developed by David Varadi of http://cssanalytics.wordpress.com/

This is the batch-calculation (parallel) version of the DV2 indicator.
*/
func DV2(h, l, c []float64, period, maLen int) []float64 {
	n := len(c)
	res := make([]float64, n)
	if n == 0 {
		return res
	}

	// Step 1: Calculate chl = c/((h+l)/2) - 1.
	chls := make([]float64, 0, maLen)
	sumChl := 0.0
	dv := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		if math.IsNaN(h[i]) || math.IsNaN(l[i]) || math.IsNaN(c[i]) {
			res[i] = math.NaN()
			continue
		}
		den := (h[i] + l[i]) / 2
		if den == 0 {
			res[i] = math.NaN()
			continue
		}
		chlVal := c[i]/den - 1
		chls = append(chls, chlVal)
		sumChl += chlVal
		if len(chls) < maLen {
			dv = append(dv, 0)
			res[i] = math.NaN()
			continue
		}
		if len(chls) > maLen {
			sumChl -= chls[0]
			chls = chls[1:]
		}
		dvVal := sumChl / float64(maLen)
		dv = append(dv, dvVal)
		if len(dv) <= period {
			res[i] = math.NaN()
			continue
		}

		lowNum, equalNum := 0.0, 0.0
		vals := dv[len(dv)-period:]
		for j := 0; j < period; j++ {
			val := vals[j]
			if val < dvVal {
				lowNum += 1
			} else if val == dvVal {
				equalNum += 1
			}
		}

		hitNum := lowNum + (equalNum+1)/2
		res[i] = hitNum * 100 / float64(period)
	}
	return res
}

// UTBot UT Bot Alerts 并行计算版本
func UTBot(c, atr []float64, rate float64) []float64 {
	n := len(c)
	if n == 0 || len(atr) != n {
		return make([]float64, n)
	}

	// 预计算 nLoss 数组
	nLoss := make([]float64, n)
	for i := 0; i < n; i++ {
		if math.IsNaN(atr[i]) {
			nLoss[i] = math.NaN()
		} else {
			nLoss[i] = atr[i] * rate
		}
	}

	signals := make([]float64, n)
	trailingStops := make([]float64, n)

	// 找到第一个有效的价格索引
	firstValid := findFirstValidIndex(c)
	for i := 0; i < firstValid; i++ {
		signals[i] = math.NaN()
	}
	if firstValid == n {
		return signals // 全是NaN
	}

	// 初始化第一个有效值
	signals[firstValid] = math.NaN()
	if !math.IsNaN(nLoss[firstValid]) {
		trailingStops[firstValid] = c[firstValid] - nLoss[firstValid]
	}

	// 计算其余值
	for i := firstValid + 1; i < n; i++ {
		prevPrice := c[i-1]
		currentPrice := c[i]
		prevStop := trailingStops[i-1]

		if math.IsNaN(nLoss[i]) {
			trailingStops[i] = prevStop
			signals[i] = math.NaN()
			continue
		}

		// 计算新的止损线
		newStop := calculateTrailingStop(currentPrice, prevPrice, prevStop, nLoss[i])
		trailingStops[i] = newStop

		// 计算信号
		signals[i] = calculateSignal(currentPrice, prevPrice, newStop, prevStop)
	}

	return signals
}

// calculateTrailingStop 计算动态止损线
func calculateTrailingStop(price, prevPrice, prevStop, nLoss float64) float64 {
	if prevStop == 0 { // 初始状态
		return price - nLoss
	}

	// 根据价格与前一止损线的关系动态调整止损位
	if price > prevStop && prevPrice > prevStop {
		// 价格上涨且持续高于止损线时，上移止损
		return math.Max(prevStop, price-nLoss)
	} else if price < prevStop && prevPrice < prevStop {
		// 价格下跌且持续低于止损线时，下移止损
		return math.Min(prevStop, price+nLoss)
	} else {
		// 价格反向突破时，重置止损
		if price > prevStop {
			return price - nLoss
		} else {
			return price + nLoss
		}
	}
}

// calculateSignal 计算交易信号
func calculateSignal(price, prevPrice, stop, prevStop float64) float64 {
	// 信号判断
	above := prevPrice <= prevStop && price > stop
	below := prevPrice >= prevStop && price < stop

	if price > stop && above {
		return 1 // 买入信号
	} else if price < stop && below {
		return -1 // 卖出信号
	} else {
		return 0 // 无信号
	}
}

// findFirstValidIndex 找到第一个非NaN值的索引
func findFirstValidIndex(data []float64) int {
	for i, v := range data {
		if !math.IsNaN(v) {
			return i
		}
	}
	return len(data)
}

// calcHLRangePct 在历史窗口中计算当前值的百分比位置
func calcHLRangePct(his []float64, cur float64) float64 {
	if len(his) == 0 {
		return 0
	}
	minVal, maxVal := math.MaxFloat64, -math.MaxFloat64
	hasVal := false
	for _, val := range his {
		if math.IsNaN(val) {
			continue
		}
		hasVal = true
		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
	}
	if !hasVal {
		return 0
	}
	rangeSize := maxVal - minVal
	if rangeSize > 0 {
		return (cur - minVal) / rangeSize * 100
	}
	return 0
}

// STC 计算Schaff趋势周期指标 (并行计算版本)
func STC(data []float64, period, fast, slow int, alpha float64) []float64 {
	n := len(data)
	if n == 0 {
		return []float64{}
	}

	// 1. 预先计算整个MACD序列
	fastEMAs := EMABy(data, fast, 0)
	slowEMAs := EMABy(data, slow, 0)
	macds := make([]float64, n)
	for i := 0; i < n; i++ {
		macds[i] = fastEMAs[i] - slowEMAs[i]
	}

	stcs := make([]float64, n)
	ddds := make([]float64, n)
	prevDDD := math.NaN()
	prevSTC := math.NaN()
	macdWindow := make([]float64, 0, period)
	dddWindow := make([]float64, 0, period)

	for i := 0; i < n; i++ {
		macd := macds[i]

		// 当MACD无效时，重置所有状态
		if math.IsNaN(macd) {
			stcs[i] = math.NaN()
			ddds[i] = math.NaN()
			prevDDD = math.NaN()
			prevSTC = math.NaN()
			macdWindow = macdWindow[:0]
			dddWindow = dddWindow[:0]
			continue
		}

		// 2. 维护MACD窗口
		macdWindow = append(macdWindow, macd)
		if len(macdWindow) > period {
			macdWindow = macdWindow[1:]
		}

		// 3. 计算第一层百分比
		ccccc := calcHLRangePct(macdWindow, macd)

		// 4. 计算第一层平滑值
		ddd := ccccc
		if !math.IsNaN(prevDDD) {
			ddd = prevDDD + alpha*(ccccc-prevDDD)
		}
		ddds[i] = ddd
		prevDDD = ddd

		// 5. 维护DDD窗口
		dddWindow = append(dddWindow, ddd)
		if len(dddWindow) > period {
			dddWindow = dddWindow[1:]
		}

		// 6. 计算第二层百分比
		dddddd := calcHLRangePct(dddWindow, ddd)

		// 7. 计算最终STC值
		stc := dddddd
		if !math.IsNaN(prevSTC) {
			stc = prevSTC + alpha*(dddddd-prevSTC)
		}
		stcs[i] = stc
		prevSTC = stc
	}

	return stcs
}

// HeikinAshi 并行版本，返回 [open, high, low, close] 数组
// 参数：开盘价、最高价、最低价、收盘价数组
func HeikinAshi(open, high, low, close []float64) ([]float64, []float64, []float64, []float64) {
	n := len(open)
	if n == 0 || len(high) != n || len(low) != n || len(close) != n {
		return nil, nil, nil, nil
	}

	hOpen := make([]float64, n)
	hHigh := make([]float64, n)
	hLow := make([]float64, n)
	hClose := make([]float64, n)

	pho := math.NaN()
	for i := 0; i < n; i++ {
		// Heikin Ashi Close 不依赖于前一个状态，可直接计算
		hClose[i] = (open[i] + high[i] + low[i] + close[i]) / 4

		if math.IsNaN(pho) {
			// 第一个有效蜡烛的开盘价
			hOpen[i] = (open[i] + close[i]) / 2
		} else {
			// 后续有效蜡烛的开盘价
			hOpen[i] = (hOpen[i-1] + hClose[i-1]) / 2
		}
		pho = hOpen[i]

		// Heikin Ashi High
		hHigh[i] = max(high[i], pho, hClose[i])

		// Heikin Ashi Low
		hLow[i] = min(low[i], pho, hClose[i])
	}

	return hOpen, hHigh, hLow, hClose
}

/*
Cross 计算两个序列在每个时间点的交叉状态。
返回值：正数表示上穿，负数表示下穿，0表示无交叉或未知。
返回值的绝对值减1 (abs(ret) - 1) 代表了最近一次交叉点到当前元素的距离。
*/
func Cross(data1 []float64, data2 []float64) []int {
	n := len(data1)
	res := make([]int, n)

	// 维护交叉状态的局部变量
	var curSign int           // 最近一次交叉的方向
	var lastIndex = -1        // 最近一次交叉点的索引
	var prevDiff = math.NaN() // 上一个有效点的差值

	for i, v1 := range data1 {
		currentDiff := v1 - data2[i]
		if math.IsNaN(currentDiff) || currentDiff == 0 {
			res[i] = curSign * (i - lastIndex + 1)
			continue
		}

		if !math.IsNaN(prevDiff) {
			if prevDiff*currentDiff < 0 {
				curSign = 1
				if currentDiff < 0 {
					curSign = -1
				}
				lastIndex = i
			}
		}

		res[i] = curSign * (i - lastIndex + 1)
		prevDiff = currentDiff
	}
	return res
}

const thresFloat64Eq = 1e-9

/*
equalNearly 判断两个float是否近似相等，解决浮点精读导致不等
*/
func equalNearly(a, b float64) bool {
	return equalIn(a, b, thresFloat64Eq)
}

/*
equalIn 判断两个float是否在一定范围内近似相等
*/
func equalIn(a, b, thres float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsInf(a, 1) && math.IsInf(b, 1) {
		return true
	}
	if math.IsInf(a, -1) && math.IsInf(b, -1) {
		return true
	}
	return math.Abs(a-b) <= thres
}
