package tav

import "math"

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

/*
equalNearly 判断两个float是否近似相等，解决浮点精读导致不等
*/
func equalNearly(a, b float64) bool {
	return equalIn(a, b, thresFloat64Eq)
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

// findFirstValidIndex 找到第一个非NaN值的索引
func findFirstValidIndex(data []float64) int {
	for i, v := range data {
		if !math.IsNaN(v) {
			return i
		}
	}
	return len(data)
}

// precomputeLinRegConstants 预计算线性回归的常数部分
func precomputeLinRegConstants(period int) (sumX, sumX2, periodF float64) {
	periodF = float64(period)
	sumX = periodF * float64(period+1) * 0.5
	sumX2 = sumX * (2*periodF + 1) / 3
	return sumX, sumX2, periodF
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

const thresFloat64Eq = 1e-9

func computeFirstSMA(data []float64, period int, index int) float64 {
	if index+1 < period {
		return math.NaN()
	}
	sum := 0.0
	count := 0
	for i := max(0, index-period+1); i <= index; i++ {
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

func maxMin(a []float64, l, r int) (float64, float64) {
	u, d := a[l], a[l]
	for i := l + 1; i <= r; i++ {
		if a[i] > u {
			u = a[i]
		}
		if a[i] < d {
			d = a[i]
		}
	}
	return u, d
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

func moveToFront(data []float64, i int) []float64 {
	res := make([]float64, 0, len(data))
	res = append(res, data[i])
	res = append(res, data[:i]...)
	if i+1 < len(data) {
		res = append(res, data[i+1:]...)
	}
	return res
}

func nan5(n int) ([]float64, []float64, []float64, []float64, []float64) {
	r := make([][]float64, 5)
	for j := range r {
		r[j] = make([]float64, n)
		for i := range r[j] {
			r[j][i] = math.NaN()
		}
	}
	return r[0], r[1], r[2], r[3], r[4]
}

func sumN(a []float64, i, n int) float64 {
	s := 0.
	for j := i - n + 1; j <= i; j++ {
		s += a[j]
	}
	return s
}
