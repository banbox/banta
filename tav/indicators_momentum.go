package tav

import "fmt"
import "math"

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

// AO is the Awesome Oscillator: SMA(HL2, fast) - SMA(HL2, slow).
func AO(high, low []float64, fast, slow int) []float64 {
	r := make([]float64, len(high))
	if fast < 1 || slow < 1 {
		for i := range r {
			r[i] = math.NaN()
		}
		return r
	}
	price := HL2(high, low)
	f := SMA(price, fast)
	s := SMA(price, slow)
	for i := range r {
		r[i] = f[i] - s[i]
	}
	return r
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

// CRSI Connors RSI，对应 stateful 指标实现中的 CRSI
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

// DPO is pandas-ta's centered=False detrended price oscillator.
func DPO(data []float64, period int) []float64 {
	r := make([]float64, len(data))
	for i := range r {
		r[i] = math.NaN()
	}
	if period < 1 {
		return r
	}
	shift := period/2 + 1
	ma := SMA(data, period)
	for i := range data {
		if i >= shift+period-1 && !math.IsNaN(ma[i-shift]) && !math.IsNaN(data[i]) {
			r[i] = data[i] - ma[i-shift]
		}
	}
	return r
}

// Dpo preserves the historical mixed-case spelling.
func Dpo(data []float64, period int) []float64 { return DPO(data, period) }

// KDJ 随机指标KDJ，对应 stateful 指标实现中的 KDJ
func KDJ(high, low, close []float64, period, sm1, sm2 int) ([]float64, []float64, []float64) {
	return KDJBy(high, low, close, period, sm1, sm2, "rma")
}

// MOM computes close momentum, matching TA-Lib MOM.
func MOM(data []float64, period int) []float64 {
	r := make([]float64, len(data))
	for i := range data {
		if i < period {
			r[i] = math.NaN()
		} else {
			r[i] = data[i] - data[i-period]
		}
	}
	return r
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

// ROC 变化率指标，对应 stateful 指标实现中的 ROC
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

// RSI 相对强度指数，对应 stateful 指标实现中的 RSI
func RSI(data []float64, period int) []float64 {
	return RSIBy(data, period, 0)
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

// StochRSI StochasticRSI，对应 stateful 指标实现中的 StochRSI
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

// TD Tom DeMark Sequence 狄马克序列，对应 stateful 指标实现中的TD
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

// UpDown 上下变动指标，对应 stateful 指标实现中的 UpDown
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

func Fisher(high, low []float64, period int) []float64 {
	o := make([]float64, len(high))
	for i := range o {
		o[i] = math.NaN()
	}
	if period < 1 {
		return o
	}
	for i := period - 1; i < len(high); i++ {
		u, _ := maxMin(high, i-period+1, i)
		_, d := maxMin(low, i-period+1, i)
		if u == d {
			continue
		}
		x := 2*((high[i]+low[i])/2-d)/(u-d) - 1
		if x > .999 {
			x = .999
		}
		if x < -.999 {
			x = -.999
		}
		o[i] = .5 * math.Log((1+x)/(1-x))
	}
	return o
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

func STOCHF(close, high, low []float64, period int, smooth ...int) ([]float64, []float64) {
	return StochF(high, low, close, period, smooth...)
}

func StochF(high, low, close []float64, period int, smooth ...int) ([]float64, []float64) {
	k := Stoch(high, low, close, period)
	raw := append([]float64(nil), k...)
	for i := 0; i <= period && i < len(k); i++ {
		k[i] = math.NaN()
	}
	d := 3
	if len(smooth) > 0 && smooth[0] > 0 {
		d = smooth[0]
	}
	return k, SMA(raw, d)
}

func ULTOSC(high, low, close []float64, short, medium, long int) []float64 {
	n := len(close)
	r := make([]float64, n)
	bp := make([]float64, n)
	tr := make([]float64, n)
	for i := 0; i < n; i++ {
		p := close[i]
		if i > 0 {
			p = close[i-1]
		}
		bp[i] = close[i] - math.Min(low[i], p)
		tr[i] = math.Max(high[i], p) - math.Min(low[i], p)
		if i < long {
			r[i] = math.NaN()
			continue
		}
		a, b, c := sumN(bp, i, short), sumN(bp, i, medium), sumN(bp, i, long)
		x, y, z := sumN(tr, i, short), sumN(tr, i, medium), sumN(tr, i, long)
		r[i] = 100 * (4*a/x + 2*b/y + c/z) / 7
	}
	return r
}

func WilliamsPercent(high, low, close []float64, period int) []float64 {
	u, _, d := Donchian(high, low, period)
	o := make([]float64, len(close))
	for i := range o {
		o[i] = math.NaN()
		if !math.IsNaN(u[i]) && u[i] != d[i] {
			o[i] = (u[i] - close[i]) / (u[i] - d[i]) * -100
		}
	}
	return o
}

// CCI 商品通道指数，对应 stateful 指标实现中的 CCI
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

func ROCR(data []float64, period int) []float64 {
	r := make([]float64, len(data))
	for i := range r {
		if period < 1 || i < period || data[i-period] == 0 {
			r[i] = math.NaN()
		} else {
			r[i] = data[i] / data[i-period]
		}
	}
	return r
}

func TRIX(data []float64, period int) []float64 {
	r := make([]float64, len(data))
	for i := range r {
		r[i] = math.NaN()
	}
	if period < 1 {
		return r
	}
	e := EMA(EMA(EMA(data, period), period), period)
	for i := 1; i < len(e); i++ {
		if !math.IsNaN(e[i]) && !math.IsNaN(e[i-1]) && e[i-1] != 0 {
			r[i] = (e[i] - e[i-1]) / e[i-1] * 100
		}
	}
	return r
}

func TSI(data []float64, short, long int) []float64 {
	r := make([]float64, len(data))
	for i := range r {
		r[i] = math.NaN()
	}
	if short < 1 || long < 1 {
		return r
	}
	m := make([]float64, len(data))
	a := make([]float64, len(data))
	for i := range m {
		m[i], a[i] = math.NaN(), math.NaN()
	}
	for i := 1; i < len(data); i++ {
		if !math.IsNaN(data[i]) && !math.IsNaN(data[i-1]) {
			m[i] = data[i] - data[i-1]
			a[i] = math.Abs(m[i])
		}
	}
	n, d := EMA(EMA(m, short), long), EMA(EMA(a, short), long)
	for i := range r {
		if !math.IsNaN(n[i]) && !math.IsNaN(d[i]) && d[i] != 0 {
			r[i] = n[i] / d[i] * 100
		}
	}
	return r
}
