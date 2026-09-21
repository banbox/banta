package banta

import "github.com/banbox/banta/tav"
import "math"

/*
	ADXBy Average Directional Index

method=0 classic ADX
method=1 TradingView "ADX and DI for v4"

suggest period: 14
smoothing: 0 to use period

return maDX
*/
func ADXBy(high *Series, low *Series, close *Series, period, smoothing, method int) *Series {
	plusDI, minusDI := pluMinDIBy(high, low, close, period, method)

	if smoothing == 0 {
		smoothing = period
	}
	// 初始化相关的系列
	dx := plusDI.To("_dx", smoothing*10+method)
	adx := plusDI.To("_adx", smoothing*10+method)
	if adx.Cached() {
		return adx
	}

	if !adx.Cached() {
		plusDIVal := plusDI.Get(0)
		if math.IsNaN(plusDIVal) {
			dx.Append(math.NaN())
			adx.Append(math.NaN())
		} else {
			minusDIVal := minusDI.Get(0)
			dx.Append(math.Abs(plusDIVal-minusDIVal) / (plusDIVal + minusDIVal) * 100)

			var maDX float64
			if method == 0 {
				maDX = RMA(dx, smoothing).Get(0)
			} else {
				maDX = SMA(dx, smoothing).Get(0)
			}
			adx.Append(maDX)
		}
	}
	return adx
}

/*
ADX Average Directional Index

suggest period: 14

return maDX
*/
func ADX(high *Series, low *Series, close *Series, period int, smoothing ...int) *Series {
	if len(smoothing) > 0 {
		return ADXBy(high, low, close, period, smoothing[0], 0)
	}
	return ADXBy(high, low, close, period, 0, 0)
}

/*
Aroon 阿隆指标

Reflects the distance between the highest price and the lowest price within a certain period of time.
反映了一段时间内出现最高价和最低价距离当前时间的远近。

AroonUp: (period - HighestBar(high, period+1)) / period * 100

AroonDn: (period - LowestBar(low, period+1)) / period * 100

Osc: AroonUp - AroonDn

return [AroonUp, Osc, AroonDn]
*/
func Aroon(high *Series, low *Series, period int) (*Series, *Series, *Series) {
	res := high.To("_aroon", period)
	if !res.Cached() {
		if !res.Cached() {
			fac := -100 / float64(period)
			up := HighestBar(high, period+1).Mul(fac).Add(100)
			dn := LowestBar(low, period+1).Mul(fac).Add(100)
			osc := up.Sub(dn)
			res.Append([]*Series{up, osc, dn})
		}
	}
	return res, res.Cols[0], res.Cols[1]
}

/*
MACD

Internationally, init_type=0 is used, while MyTT and China mainly use init_type=1
国外主流使用init_type=0，MyTT和国内主要使用init_type=1

fast: 12, slow: 26, smooth: 9

return [macd, signal]
*/
func MACD(obj *Series, fast int, slow int, smooth int) (*Series, *Series) {
	return MACDBy(obj, fast, slow, smooth, 0)
}

/*
PluMinDI

suggest period: 14

return [plus di, minus di]
*/
func PluMinDI(high *Series, low *Series, close *Series, period int) (*Series, *Series) {
	return pluMinDIBy(high, low, close, period, 0)
}

/*
PluMinDM

suggest period: 14

return [Plus DM, Minus DM]
*/
func PluMinDM(high *Series, low *Series, close *Series, period int) (*Series, *Series) {
	return pluMinDMBy(high, low, close, period, 0)
}

/*
method=0 classic, use period as initLen
method=1 use period+1 as initLen, for TradingView "ADX and DI for v4"

return [Plus DM, Minus DM]
*/
func pluMinDMBy(high *Series, low *Series, close *Series, period, method int) (*Series, *Series) {
	// 初始化相关的系列
	res := close.To("_PluMinDM", period*10+method)
	if res.Cached() {
		return res, res.Cols[0]
	}
	if !res.Cached() {
		// 计算 DMH 和 DML
		dmhVal := high.Get(0) - high.Get(1)
		dmlVal := low.Get(1) - low.Get(0)
		plusDM, minusDM := 0.0, 0.0
		if dmhVal > max(dmlVal, 0) {
			plusDM = dmhVal
		} else if dmlVal > max(dmhVal, 0) {
			minusDM = dmlVal
		}

		// 计算 TR
		tr := TR(high, low, close).Get(0)
		state, _ := res.More.(*dmState)
		if state == nil {
			state = &dmState{}
			res.More = state
			res.DupMore = func(more interface{}) interface{} {
				m := more.(*dmState)
				return &dmState{m.Num, m.DmPosMA, m.DmNegMA, m.TRMA}
			}
		}
		if math.IsNaN(tr) {
			res.Append([]float64{math.NaN(), math.NaN()})
		} else {
			state.Num += 1

			// calc Wilder's smoothing of DmH/DmL/TR
			alpha := 1 / float64(period)
			initLen := period
			if method == 1 {
				initLen = period + 1
			}
			if state.Num <= initLen-1 {
				state.DmPosMA += plusDM
				state.DmNegMA += minusDM
				state.TRMA += tr
				if state.Num <= period-1 {
					res.Append([]float64{math.NaN(), math.NaN()})
					return res, res.Cols[0]
				}
			} else {
				state.DmPosMA = state.DmPosMA*(1-alpha) + plusDM
				state.DmNegMA = state.DmNegMA*(1-alpha) + minusDM
				state.TRMA = state.TRMA*(1-alpha) + tr
			}
			res.Append([]float64{state.DmPosMA, state.DmNegMA})
		}
	}
	return res, res.Cols[0]
}

// AroonOsc returns Aroon up minus down (TA-Lib AROONOSC).
func AroonOsc(high, low *Series, period int) *Series {
	// Compute the oscillator directly from the streaming window.  Calling the
	// legacy multi-column Aroon helper here makes its column cache one bar out
	// of phase with the vector implementation.
	res := high.To("_aroonosc", period)
	if !res.Cached() {
		if period <= 0 || high.Len() < period+1 {
			res.Append(math.NaN())
		} else {
			hi, lo := high.Get(0), low.Get(0)
			hiOffset, loOffset := 0, 0
			for i := 1; i <= period; i++ {
				if high.Get(i) > hi {
					hi, hiOffset = high.Get(i), i
				}
				if low.Get(i) < lo {
					lo, loOffset = low.Get(i), i
				}
			}
			res.Append(-100*float64(hiOffset)/float64(period) + 100 -
				(-100*float64(loOffset)/float64(period) + 100))
		}
	}
	return res
}

// DMI returns +DI, -DI and ADX using Wilder smoothing. The second period is
// accepted for compatibility with generated strategies and controls ADX
// smoothing while DI uses period.
func DMI(high, low, close *Series, period int, smoothing ...int) (*Series, *Series, *Series) {
	plus, minus := pluMinDIBy(high, low, close, period, 0)
	adxSmoothing := 0
	if len(smoothing) > 0 {
		adxSmoothing = smoothing[0]
	}
	return plus, minus, ADXBy(high, low, close, period, adxSmoothing, 0)
}

// Parabolic SAR. AF starts at .02, increments by .02, capped at .20.
func SAR(high, low *Series, step, max float64) *Series {
	res := high.To("_sar", int(step*1000)*1000+int(max*1000))
	if res.Cached() {
		return res
	}
	type state struct {
		sar, ep, af  float64
		up           bool
		init         bool
		bars         int
		prevH, prevL float64
	}
	s, _ := res.More.(*state)
	if s == nil {
		s = &state{up: true}
		res.More = s
		res.DupMore = func(m interface{}) interface{} { x := m.(*state); y := *x; return &y }
	}
	h, l := high.Get(0), low.Get(0)
	if !s.init {
		s.sar = l
		s.ep = h
		s.af = step
		s.prevH = h
		s.prevL = l
		s.init = true
		s.bars = 1
		res.Append(math.NaN())
		return res
	}
	if s.bars == 1 {
		// TA-Lib seeds the second bar from the current low for an initial
		// uptrend, matching the vector implementation.
		s.sar = l
		s.ep = h
		s.af = step
		s.prevH, s.prevL = h, l
		s.bars = 2
		res.Append(s.sar)
		return res
	}
	s.sar = s.sar + s.af*(s.ep-s.sar)
	if s.up {
		s.sar = math.Min(s.sar, s.prevL)
		if s.bars > 2 {
			s.sar = math.Min(s.sar, low.Get(1))
		}
		if s.bars > 2 {
			s.sar = math.Min(s.sar, low.Get(2))
		}
		if h > s.ep {
			s.ep = h
			s.af = math.Min(max, s.af+step)
		}
		if l < s.sar {
			s.up = false
			s.sar = s.ep
			s.ep = l
			s.af = step
		}
	} else {
		s.sar = math.Max(s.sar, s.prevH)
		if s.bars > 2 {
			s.sar = math.Max(s.sar, high.Get(1))
		}
		if s.bars > 2 {
			s.sar = math.Max(s.sar, high.Get(2))
		}
		if l < s.ep {
			s.ep = l
			s.af = math.Min(max, s.af+step)
		}
		if h > s.sar {
			s.up = true
			s.sar = s.ep
			s.ep = h
			s.af = step
		}
	}
	s.prevH = h
	s.prevL = l
	s.bars++
	res.Append(s.sar)
	return res
}

func AROONOSC(high, low *Series, period int) *Series { return AroonOsc(high, low, period) }

func Ichimoku(high, low, close *Series, conversion, base, span int) (*Series, *Series, *Series, *Series, *Series) {
	r := close.To("_ichimoku", conversion*1000000+base*1000+span)
	if !r.Cached() {
		a, b, c, d, e := tav.Ichimoku(hist(high), hist(low), hist(close), conversion, base, span)
		r.Append([]float64{last(a), last(b), last(c), last(d), last(e)})
	}
	return r, r.Cols[0], r.Cols[1], r.Cols[2], r.Cols[3]
}

func KST(obj *Series, r1, r2, r3, r4, s1, s2, s3, s4 int) *Series {
	r := obj.To("_kst", r1*1e7+r2*1e6+r3*1e5+r4*1e4+s1*1e3+s2*100+s3*10+s4)
	if !r.Cached() {
		r.Append(last(tav.KST(hist(obj), r1, r2, r3, r4, s1, s2, s3, s4)))
	}
	return r
}

func MACDBy(obj *Series, fast int, slow int, smooth int, initType int) (*Series, *Series) {
	res := obj.To("_macd", fast*1000+slow*100+smooth*10+initType)
	if !res.Cached() {
		if !res.Cached() {
			short := EMABy(obj, fast, initType)
			longMA := EMABy(obj, slow, initType)
			macd := short.Sub(longMA)
			signal := EMABy(macd, smooth, initType)
			res.Append([]float64{macd.Get(0), signal.Get(0)})
		}
	}
	return res, res.Cols[0]
}

func PSAR(high, low *Series, step, max float64) *Series { return SAR(high, low, step, max) }

func pluMinDIBy(high *Series, low *Series, close *Series, period, method int) (*Series, *Series) {
	plusDM, _ := pluMinDMBy(high, low, close, period, method)
	res := plusDM.To("_PluMinDI", period*10+method)
	if !res.Cached() {
		if !res.Cached() {
			plusDmVal := plusDM.Get(0)
			if math.IsNaN(plusDmVal) {
				res.Append([]float64{math.NaN(), math.NaN()})
			} else {
				// calc dx
				state, _ := plusDM.More.(*dmState)
				plusDI := 100 * state.DmPosMA / state.TRMA
				minusDI := 100 * state.DmNegMA / state.TRMA
				res.Append([]float64{plusDI, minusDI})
			}
		}
	}

	return res, res.Cols[0]
}

/*
UTBot UT Bot Alerts from TradingView
*/
func UTBot(c, atr *Series, rate float64) *Series {
	res := atr.To("_utBot", int(rate*10))
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		prevXATRTrailingStop, _ := res.More.(float64)
		nLoss := atr.Mul(rate).Get(0)
		// 计算动态止损线
		price := c.Get(0)
		prevSrc := c.Get(1)
		if math.IsNaN(nLoss) {
			res.Append(math.NaN())
		} else {
			var xATRTrailingStop float64
			if prevXATRTrailingStop == 0 { // 初始状态
				xATRTrailingStop = price - nLoss
			} else {
				//根据价格与前一止损线的关系动态调整止损位：
				prevStop := prevXATRTrailingStop
				if price > prevStop && prevSrc > prevStop {
					//价格上涨且持续高于止损线时，上移止损。
					xATRTrailingStop = math.Max(prevStop, price-nLoss)
				} else if price < prevStop && prevSrc < prevStop {
					//价格下跌且持续低于止损线时，下移止损。
					xATRTrailingStop = math.Min(prevStop, price+nLoss)
				} else {
					//价格反向突破时，重置止损。
					if price > prevStop {
						xATRTrailingStop = price - nLoss
					} else {
						xATRTrailingStop = price + nLoss
					}
				}
			}

			// 信号判断
			above := prevSrc <= prevXATRTrailingStop && price > xATRTrailingStop
			below := prevSrc >= prevXATRTrailingStop && price < xATRTrailingStop
			// 更新状态
			res.More = xATRTrailingStop

			if price > xATRTrailingStop && above {
				res.Append(1)
			} else if price < xATRTrailingStop && below {
				res.Append(-1)
			} else {
				res.Append(0)
			}
		}
	}
	return res
}

func DX(high, low, close *Series, period int) *Series {
	r := close.To("_dx", period)
	if !r.Cached() {
		r.Append(last(tav.DX(hist(high), hist(low), hist(close), period)))
	}
	return r
}
