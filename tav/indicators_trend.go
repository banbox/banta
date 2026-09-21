package tav

import "math"

// ADX 平均方向指数，对应 stateful 指标实现中的 ADX
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

// Aroon 阿隆指标，对应 stateful 指标实现中的 Aroon
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

// MACD 移动平均线收敛发散指标，对应 stateful 指标实现中的 MACD
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

// SAR computes Parabolic SAR with the canonical Wilder acceleration rules.
func SAR(high, low []float64, step, max float64) []float64 {
	n := len(high)
	r := make([]float64, n)
	if n == 0 {
		return r
	}
	up := true
	sar := low[0]
	ep := high[0]
	af := step
	r[0] = math.NaN()
	if n == 1 {
		return r
	}
	// TA-Lib seeds the second bar from that bar's low for an initial uptrend.
	sar = low[1]
	r[1] = sar
	ep = high[1]
	for i := 2; i < n; i++ {
		sar = sar + af*(ep-sar)
		if up {
			sar = math.Min(sar, low[i-1])
			if i > 2 {
				sar = math.Min(sar, low[i-2])
			}
			if high[i] > ep {
				ep = high[i]
				af = math.Min(max, af+step)
			}
			if low[i] < sar {
				up = false
				sar = ep
				ep = low[i]
				af = step
			}
		} else {
			sar = math.Max(sar, high[i-1])
			if i > 2 {
				sar = math.Max(sar, high[i-2])
			}
			if low[i] < ep {
				ep = low[i]
				af = math.Min(max, af+step)
			}
			if high[i] > sar {
				up = true
				sar = ep
				ep = high[i]
				af = step
			}
		}
		r[i] = sar
	}
	return r
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

func AROONOSC(high, low []float64, period int) []float64 { return AroonOsc(high, low, period) }

func AroonOsc(high, low []float64, period int) []float64 {
	_, o, _ := Aroon(high, low, period)
	return o
}

func Ichimoku(high, low, close []float64, conversion, base, span int) ([]float64, []float64, []float64, []float64, []float64) {
	n := len(close)
	conv, bas, a, b, lag := nan5(n)
	for i := 0; i < n; i++ {
		if i+1 >= conversion {
			u, _ := maxMin(high, i-conversion+1, i)
			_, d := maxMin(low, i-conversion+1, i)
			conv[i] = (u + d) / 2
		}
		if i+1 >= base {
			u, _ := maxMin(high, i-base+1, i)
			_, d := maxMin(low, i-base+1, i)
			bas[i] = (u + d) / 2
		}
		if !math.IsNaN(conv[i]) && !math.IsNaN(bas[i]) {
			a[i] = (conv[i] + bas[i]) / 2
		}
		if i+1 >= span {
			u, _ := maxMin(high, i-span+1, i)
			_, d := maxMin(low, i-span+1, i)
			b[i] = (u + d) / 2
		}
		if i >= base {
			lag[i] = close[i-base]
		}
	}
	return conv, bas, a, b, lag
}

func KST(data []float64, r1, r2, r3, r4, s1, s2, s3, s4 int) []float64 {
	roc := func(p int) []float64 {
		r := make([]float64, len(data))
		for i := range r {
			r[i] = math.NaN()
			if i >= p && data[i-p] != 0 {
				r[i] = (data[i] - data[i-p]) / data[i-p] * 100
			}
		}
		return r
	}
	a, b, c, d := roc(r1), roc(r2), roc(r3), roc(r4)
	x, y, z, w := SMA(a, s1), SMA(b, s2), SMA(c, s3), SMA(d, s4)
	o := make([]float64, len(data))
	for i := range o {
		o[i] = x[i] + 2*y[i] + 3*z[i] + 4*w[i]
	}
	return o
}

func PSAR(high, low []float64, step, max float64) []float64 { return SAR(high, low, step, max) }

func PluMinDI(high, low, close []float64, period int) ([]float64, []float64) {
	return pluMinDIBy(high, low, close, period, 0)
}

func PluMinDM(high, low, cls []float64, period int) ([]float64, []float64) {
	plus, minu, _ := pluMinDMBy(high, low, cls, period, 0)
	return plus, minu
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

func DX(high, low, close []float64, period int) []float64 {
	n := len(close)
	o := make([]float64, n)
	for i := range o {
		o[i] = math.NaN()
	}
	if period < 1 {
		return o
	}
	for i := period; i < n; i++ {
		up := math.Max(high[i]-high[i-1], 0)
		dn := math.Max(low[i-1]-low[i], 0)
		tr := math.Max(high[i]-low[i], math.Max(math.Abs(high[i]-close[i-1]), math.Abs(low[i]-close[i-1])))
		if tr != 0 {
			o[i] = math.Abs(up-dn) / (up + dn) * 100
		}
	}
	return o
}
