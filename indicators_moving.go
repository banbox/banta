package banta

import "github.com/banbox/banta/tav"
import "math"

/*
ALMA Arnaud Legoux Moving Average

period:  window size. Default: 10

sigma: Smoothing value. Default 6.0

distOff: min 0 (smoother), max 1 (more responsive). Default: 0.85
*/
func ALMA(obj *Series, period int, sigma, distOff float64) *Series {
	res := obj.To("_alma", pkey(ikey(period), fkey(sigma), fkey(distOff)))
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		inVal := obj.Get(0)
		if math.IsNaN(inVal) {
			res.Append(math.NaN())
		} else {
			arr := WrapFloatArr(res, period, inVal)
			if len(arr) < period {
				res.Append(math.NaN())
			} else {
				m := distOff * (float64(period) - 1)
				s := float64(period) / sigma
				var windowSum, cumSum = float64(0), float64(0)
				for i := range arr {
					fi := float64(i)
					wei := math.Exp(-(fi - m) * (fi - m) / (2 * s * s))
					windowSum += wei * arr[period-i-1]
					cumSum += wei
				}
				res.Append(windowSum / cumSum)
			}
		}
	}
	return res
}

/*
EMA Exponential Moving Average 指数移动均线

Latest value weight: 2/(n+1)

最近一个权重：2/(n+1)
*/
func EMA(obj *Series, period int) *Series {
	return EMABy(obj, period, 0)
}

/*
EMABy 指数移动均线
最近一个权重：2/(n+1)
initType：0使用SMA初始化，1第一个有效值初始化
*/
func EMABy(obj *Series, period int, initType int) *Series {
	res := obj.To("_ema", period*10+initType)
	alpha := 2.0 / float64(period+1)
	return ewma(obj, res, period, alpha, initType, math.NaN())
}

/*
HMA Hull Moving Average

suggest period: 9
*/
func HMA(obj *Series, period int) *Series {
	maLen := int(math.Floor(math.Sqrt(float64(period))))
	mid := obj.To("_hmamid", period)
	if mid.Cached() {
		return WMA(mid, maLen)
	}
	if !mid.Cached() {
		half := WMA(obj, period/2).Get(0)
		wma := WMA(obj, period).Get(0)
		mid.Append(2*half - wma)
	}
	return WMA(mid, maLen)
}

/*
KAMA Kaufman Adaptive Moving Average

period: 10 fixed: (fast: 2, slow: 30)
*/
func KAMA(obj *Series, period int) *Series {
	return KAMABy(obj, period, 2, 30)
}

/*
KAMABy Kaufman Adaptive Moving Average

period: 10, fast: 2, slow: 30
*/
func KAMABy(obj *Series, period int, fast, slow int) *Series {
	res := obj.To("_kama", period*10000+slow*100+fast)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		prevRes, ok := res.More.(float64)
		if !ok {
			prevRes = math.NaN()
		}
		effRatio := ER(obj, period).Get(0)

		resVal := math.NaN()
		if !math.IsNaN(effRatio) {
			fastV := 2 / float64(fast+1)
			slowV := 2 / float64(slow+1)
			alpha := math.Pow(effRatio*(fastV-slowV)+slowV, 2)
			curVal := obj.Get(0)
			if math.IsNaN(prevRes) {
				prevRes = obj.Get(1)
			}
			resVal = alpha*curVal + (1-alpha)*prevRes
			res.More = resVal
		}
		res.Append(resVal)
	}
	return res
}

/*
RMA Relative Moving Average 相对移动均线

The difference from EMA is: both the numerator and denominator are reduced by 1
Latest value weight: 1/n

	和EMA区别是：分子分母都减一
	最近一个权重：1/n
*/
func RMA(obj *Series, period int) *Series {
	return RMABy(obj, period, 0, math.NaN())
}

/*
RMABy Relative Moving Average 相对移动均线

The difference from EMA is: both the numerator and denominator are reduced by 1
The most recent weight: 1/n

	和EMA区别是：分子分母都减一
	最近一个权重：1/n

initType: 0 initialize with SMA, 1 initialize with the first valid value

initVal defaults to Nan

initType：0使用SMA初始化，1第一个有效值初始化

initVal 默认Nan
*/
func RMABy(obj *Series, period int, initType int, initVal float64) *Series {
	res := obj.To("_rma", pkey(ikey(period), ikey(initType), fkey(initVal)))
	alpha := 1.0 / float64(period)
	return ewma(obj, res, period, alpha, initType, initVal)
}

/*
VWMA Volume Weighted Moving Average 成交量加权平均价格

sum(price*volume)/sum(volume)

suggest period: 20
*/
func VWMA(price *Series, vol *Series, period int) *Series {
	res := price.To("_vwma", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		volVal := vol.Get(0)
		cost := price.Get(0) * volVal
		more, _ := res.More.(*moreVWMA)
		if more == nil {
			more = &moreVWMA{}
			res.More = more
			res.DupMore = func(mAny interface{}) interface{} {
				m := mAny.(*moreVWMA)
				return &moreVWMA{m.sumCost, m.sumWei, append([]float64{}, m.costs...), append([]float64{}, m.volumes...)}
			}
		}
		if math.IsNaN(cost) {
			res.Append(math.NaN())
		} else {
			more.sumCost += cost
			more.sumWei += volVal
			more.volumes = append(more.volumes, volVal)
			more.costs = append(more.costs, cost)
			if len(more.volumes) > period {
				more.sumCost -= more.costs[0]
				more.sumWei -= more.volumes[0]
				more.costs = more.costs[1:]
				more.volumes = more.volumes[1:]
			}
			if len(more.volumes) < period {
				res.Append(math.NaN())
			} else {
				res.Append(more.sumCost / more.sumWei)
			}
		}
	}
	return res
}

/*
WMA Weighted Moving Average.

the weighting factors decrease in arithmetic progression.

suggest period: 9
*/
func WMA(obj *Series, period int) *Series {
	res := obj.To("_wma", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		val := obj.Get(0)
		if math.IsNaN(val) {
			res.Append(math.NaN())
		} else {
			more, _ := res.More.(*wmaSta)
			if more == nil {
				more = &wmaSta{}
				res.More = more
				res.DupMore = func(mAny interface{}) interface{} {
					m := mAny.(*wmaSta)
					return &wmaSta{append([]float64{}, m.arr...), m.allSum, m.weiSum}
				}
			}
			more.arr = append(more.arr, val)
			if len(more.arr) > period {
				more.allSum -= more.weiSum
				more.weiSum -= more.arr[0]
				more.arr = more.arr[1:]
			}
			arrNum := len(more.arr)
			more.weiSum += val
			more.allSum += val * float64(arrNum)
			if arrNum < period {
				res.Append(math.NaN())
			} else {
				sumWei := float64(period) * float64(period+1) * 0.5
				res.Append(more.allSum / sumWei)
			}
		}
	}
	return res
}

/*
alpha: update weight for latest value
initType: 0: sma   1: first value
initVal: use this as init val if not nan
*/
func ewma(obj, res *Series, period int, alpha float64, initType int, initVal float64) *Series {
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		prevRes, ok := res.More.(float64)
		if !ok {
			prevRes = math.NaN()
			res.More = prevRes
		}
		inVal := obj.Get(0)
		var resVal float64
		if math.IsNaN(inVal) {
			resVal = inVal
		} else {
			if math.IsNaN(prevRes) {
				if !math.IsNaN(initVal) {
					// 使用给定值作为计算第一个值的前置值
					resVal = alpha*inVal + (1-alpha)*initVal
				} else if initType == 0 {
					// 使用 SMA 作为第一个 EMA 值
					resVal = SMA(obj, period).Get(0)
				} else {
					// 第一个有效值作为第一个 EMA 值
					resVal = inVal
				}
			} else {
				resVal = alpha*inVal + (1-alpha)*prevRes
			}
			// 如果当前计算结果有效，则更新状态
			if !math.IsNaN(resVal) {
				res.More = resVal
			}
		}
		res.Append(resVal)
	}
	return res
}

// DEMA is double exponential moving average: 2*EMA - EMA(EMA).
func DEMA(obj *Series, period int) *Series {
	res := obj.To("_dema", period)
	if !res.Cached() {
		e := EMA(obj, period)
		res.Append(2*e.Get(0) - EMA(e, period).Get(0))
	}
	return res
}

// SMMA is the common generated name for Wilder's smoothed moving average.
func SMMA(obj *Series, period int) *Series { return RMA(obj, period) }

// T3 is Tillson's T3 moving average, using the standard vFactor=.7.
func T3(obj *Series, period int) *Series {
	res := obj.To("_t3", period)
	if !res.Cached() {
		v := 0.7
		e1 := EMA(obj, period)
		e2 := EMA(e1, period)
		e3 := EMA(e2, period)
		e4 := EMA(e3, period)
		e5 := EMA(e4, period)
		e6 := EMA(e5, period)
		c1 := -v * v * v
		c2 := 3*v*v + 3*v*v*v
		c3 := -6*v*v - 3*v - 3*v*v*v
		c4 := 1 + 3*v + v*v*v + 3*v*v
		res.Append(c1*e6.Get(0) + c2*e5.Get(0) + c3*e4.Get(0) + c4*e3.Get(0))
	}
	return res
}

// TEMA is the triple exponential moving average.
func TEMA(obj *Series, period int) *Series {
	e1 := EMA(obj, period)
	e2 := EMA(e1, period)
	e3 := EMA(e2, period)
	return e1.Mul(3).Sub(e2.Mul(3)).Add(e3)
}

// VWAP returns the cumulative volume-weighted typical price for the replay
// segment. Generated strategies use it as a baseline and do not provide a
// rolling period or session-reset signal.
func VWAP(first, second *Series, rest ...*Series) *Series {
	close, volume := first, second
	var high, low *Series
	if len(rest) == 2 {
		high, low, close, volume = first, second, rest[0], rest[1]
	}
	res := close.To("_vwap", 0)
	if res.Cached() {
		return res
	}
	state, _ := res.More.(*vwapState)
	if state == nil {
		state = &vwapState{}
		res.More = state
		res.DupMore = func(more interface{}) interface{} {
			m := more.(*vwapState)
			return &vwapState{sumCost: m.sumCost, sumVolume: m.sumVolume}
		}
	}
	h, l, c, v := close.Get(0), close.Get(0), close.Get(0), volume.Get(0)
	if high != nil && low != nil {
		h, l = high.Get(0), low.Get(0)
	}
	if math.IsNaN(h) || math.IsNaN(l) || math.IsNaN(c) || math.IsNaN(v) {
		res.Append(math.NaN())
		return res
	}
	state.sumCost += ((h + l + c) / 3) * v
	state.sumVolume += v
	if state.sumVolume == 0 {
		res.Append(math.NaN())
	} else {
		res.Append(state.sumCost / state.sumVolume)
	}
	return res
}

func MAMA(obj *Series, fast, slow float64) (*Series, *Series) {
	r := obj.To("_mama", pkey(fkey(fast), fkey(slow)))
	if !r.Cached() {
		a, b := tav.MAMA(hist(obj), fast, slow)
		r.Append([]float64{last(a), last(b)})
	}
	return r, r.Cols[0]
}

func SMA(obj *Series, period int) *Series {
	res := obj.To("_sma", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		midObj := Sum(obj, period)
		if midObj.Len() >= period {
			res.Append(midObj.Get(0) / float64(period))
		} else {
			res.Append(math.NaN())
		}
	}
	return res
}

func SSF(obj *Series, period int) *Series {
	r := obj.To("_ssf", period)
	if r.Cached() {
		return r
	}
	r.Append(last(tav.SSF(hist(obj), period)))
	return r
}

func SWMA(obj *Series) *Series {
	res := obj.To("_swma", 0)
	if !res.Cached() {
		if len(obj.Data) < 4 {
			res.Append(math.NaN())
		} else {
			res.Append((obj.Get(0) + 2*obj.Get(1) + 2*obj.Get(2) + obj.Get(3)) / 6)
		}
	}
	return res
}

func Sum(obj *Series, period int) *Series {
	res := obj.To("_sum", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		sta, _ := res.More.(*sumState)
		if sta == nil {
			sta = &sumState{}
			res.More = sta
			res.DupMore = func(more interface{}) interface{} {
				s := more.(*sumState)
				return &sumState{s.sumVal, append([]float64{}, s.arr...)}
			}
		}
		curVal := obj.Get(0)
		if !math.IsNaN(curVal) {
			// 跳过nan
			sta.sumVal += curVal
			sta.arr = append(sta.arr, curVal)
			if len(sta.arr) > period {
				sta.sumVal -= sta.arr[0]
				sta.arr = sta.arr[1:]
			}
			if len(sta.arr) >= period {
				res.Append(sta.sumVal)
				return res
			}
		}
		res.Append(math.NaN())
	}
	return res
}

func TRIMA(obj *Series, period int) *Series {
	first := (period + 1) / 2
	second := period/2 + 1
	return SMA(SMA(obj, first), second)
}

func VIDYA(obj *Series, period int) *Series {
	r := obj.To("_vidya", period)
	if r.Cached() {
		return r
	}
	r.Append(last(tav.VIDYA(hist(obj), period)))
	return r
}

func ZLMA(obj *Series, period int) *Series {
	lag := (period - 1) / 2
	adj := obj.Add(obj.Sub(obj.Back(lag)))
	return EMA(adj, period)
}

/*
ER Efficiency Ratio / Trend to Noise Ratio

suggest period: 8
*/
func ER(obj *Series, period int) *Series {
	res := obj.To("_tnr", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		sta, _ := res.More.(*tnrState)
		if sta == nil {
			sta = &tnrState{
				prevIn: math.NaN(),
			}
			res.More = sta
			res.DupMore = func(more interface{}) interface{} {
				m := more.(*tnrState)
				return &tnrState{append([]float64{}, m.arr...), m.sumVal, m.prevIn, append([]float64{}, m.arrIn...)}
			}
		}
		inVal := obj.Get(0)
		curVal := math.Abs(inVal - sta.prevIn)
		if !math.IsNaN(inVal) {
			sta.prevIn = inVal
			sta.arrIn = append(sta.arrIn, inVal)
		}
		var resVal = math.NaN()
		if math.IsNaN(curVal) {
			res.Append(resVal)
		} else {
			sta.sumVal += curVal
			if len(sta.arr) < period {
				sta.arr = append(sta.arr, curVal)
			} else {
				sta.sumVal -= sta.arr[0]
				sta.arr = append(sta.arr[1:], curVal)
			}
			if len(sta.arrIn) > period {
				periodVal := sta.arrIn[0]
				sta.arrIn = sta.arrIn[1:]
				if sta.sumVal > 0 {
					diffVal := math.Abs(inVal - periodVal)
					resVal = diffVal / sta.sumVal
				}
			}
			res.Append(resVal)
		}
	}
	return res
}
