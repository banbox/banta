package banta

import "fmt"
import "github.com/banbox/banta/tav"
import "math"

/*
	TD Tom DeMark Sequence（狄马克序列）

over bought: 9,13

over sell: -9, -13

9和13表示超买；-9和-13表示超卖
*/
func TD(obj *Series) *Series {
	res := obj.To("_td", 0)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		inVal := obj.Get(0)
		if math.IsNaN(inVal) {
			res.Append(math.NaN())
		} else {
			prevs := WrapFloatArr(res, 5, inVal)
			if len(prevs) < 5 {
				res.Append(math.NaN())
			} else {
				sub4 := inVal - prevs[0]
				prevNum := res.Get(0)
				step := 1
				if equalNearly(sub4, 0) {
					step = 0
				} else if sub4 < 0 {
					step = -1
				}
				if !math.IsNaN(prevNum) && prevNum*sub4 > 0 {
					resVal := int(math.Round(prevNum)) + step
					res.Append(resVal)
				} else {
					res.Append(step)
				}
			}
		}
	}
	return res
}

/*
CMO Chande Momentum Oscillator

suggest period: 9

Same implementation as ta-lib
For TradingView, use: CMOBy(obj, period, 1)
*/
func CMO(obj *Series, period int) *Series {
	return CMOBy(obj, period, 0)
}

/*
CMOBy Chande Momentum Oscillator

suggest period: 9

maType: 0: ta-lib   1: tradingView
*/
func CMOBy(obj *Series, period int, maType int) *Series {
	res := obj.To("_cmo", period*10+maType)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		sta, _ := res.More.(*cmdSta)
		if sta == nil {
			sta = &cmdSta{
				prevIn: math.NaN(),
			}
			res.More = sta
			res.DupMore = func(more interface{}) interface{} {
				m := more.(*cmdSta)
				return &cmdSta{append([]float64{}, m.subs...), m.sumPos, m.sumNeg, m.prevIn}
			}
		}
		inVal := obj.Get(0)
		val := inVal - sta.prevIn
		if !math.IsNaN(inVal) {
			sta.prevIn = inVal
		}
		if !math.IsNaN(val) {
			if maType == 0 {
				// ta-lib  wilder's smooth
				if len(sta.subs) >= period {
					wei := 1 - 1/float64(period)
					sta.sumPos *= wei
					sta.sumNeg *= wei
					if val > 0 {
						sta.sumPos += val * (1 - wei)
					} else {
						sta.sumNeg -= val * (1 - wei)
					}
					sta.subs = append(sta.subs[1:], val)
				} else {
					if val > 0 {
						sta.sumPos += val
					} else {
						sta.sumNeg -= val
					}
					sta.subs = append(sta.subs, val)
					if len(sta.subs) == period {
						sta.sumPos /= float64(period)
						sta.sumNeg /= float64(period)
					}
				}
			} else {
				// tradingView  Sum(sub, period)
				if val > 0 {
					sta.sumPos += val
				} else {
					sta.sumNeg -= val
				}
				if len(sta.subs) >= period {
					prevVal := sta.subs[0]
					if prevVal > 0 {
						sta.sumPos -= prevVal
					} else {
						sta.sumNeg += prevVal
					}
					sta.subs = append(sta.subs[1:], val)
				} else {
					sta.subs = append(sta.subs, val)
				}
			}
			if len(sta.subs) < period {
				res.Append(math.NaN())
			} else {
				res.Append((sta.sumPos - sta.sumNeg) * 100 / (sta.sumPos + sta.sumNeg))
			}
		} else {
			res.Append(math.NaN())
		}
	}
	return res
}

/*
CRSI Connors RSI

suggest period:3, upDn:2, roc:100

Basically the same as TradingView
*/
func CRSI(obj *Series, period, upDn, roc int) *Series {
	return CRSIBy(obj, period, upDn, roc, 0)
}

/*
CRSIBy Connors RSI

suggest period:3, upDn:2, roc:100

vtype: 0 Calculation in TradingView method

1 Calculation in ta-lib community method:

	chg = close_col / close_col.shift(1)
	updown = np.where(chg.gt(1), 1.0, np.where(chg.lt(1), -1.0, 0.0))
	rsi = ta.RSI(close_arr, timeperiod=3)
	ud = ta.RSI(updown, timeperiod=2)
	roc = ta.ROC(close_arr, 100)
	crsi = (rsi + ud + roc) / 3
*/
func CRSIBy(obj *Series, period, upDn, roc, vtype int) *Series {
	res := obj.To("_crsi", roc*100000+vtype*10000+upDn*100+period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		rsi := RSI(obj, period).Get(0)
		ud := RSI(UpDown(obj, vtype), upDn).Get(0)
		var rc float64
		if vtype == 0 {
			rc = PercentRank(ROC(obj, 1), roc).Get(0)
		} else {
			rc = ROC(obj, roc).Get(0)
		}
		res.Append((rsi + ud + rc) / 3)
	}
	return res
}

/*
DV2 Developed by David Varadi of http://cssanalytics.wordpress.com/

	period: 252   maLen: 2

	This seems to be the *Bounded* version.
	这里和backtrader中实现不一致，是另一个版本。

	See also:

	  - http://web.archive.org/web/20131216100741/http://quantingdutchman.wordpress.com/2010/08/06/dv2-indicator-for-amibroker/
	  - https://www.reddit.com/r/CapitalistExploits/comments/1d0azms/david_varadis_dv2_indicator_trading_strategies/
*/
func DV2(h, l, c *Series, period, maLen int) *Series {
	res := c.To("_dv", period*100+maLen)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		sta, _ := res.More.(*dv2Sta)
		if sta == nil {
			sta = &dv2Sta{}
			res.More = sta
			res.DupMore = func(more interface{}) interface{} {
				m := more.(*dv2Sta)
				return &dv2Sta{append([]float64{}, m.chl...), append([]float64{}, m.dv...)}
			}
		}
		h0, l0, c0 := h.Get(0), l.Get(0), c.Get(0)
		// 如果当前输入值为 NaN，返回NaN然后跳过，不重置状态
		if math.IsNaN(h0) || math.IsNaN(l0) || math.IsNaN(c0) {
			res.Append(math.NaN())
		} else {
			// calc close*2/(high+low)
			den := (h0 + l0) / 2
			if den == 0 {
				res.Append(math.NaN()) // 避免除以零，视为无效值
			} else {
				chl := c0/den - 1
				sta.chl = append(sta.chl, chl)
				// apply maLen to chl
				chlLen := len(sta.chl)
				if chlLen < maLen {
					sta.dv = append(sta.dv, 0)
					res.Append(math.NaN())
				} else {
					sum := float64(0)
					for i := chlLen - maLen; i < chlLen; i++ {
						sum += sta.chl[i]
					}
					dvVal := sum / float64(maLen)
					sta.dv = append(sta.dv, dvVal)
					if chlLen > maLen*3 {
						sta.chl = sta.chl[chlLen-maLen:]
					}
					if len(sta.dv) > period {
						// percent rank for dv
						lowNum, equalNum := float64(0), float64(0)
						vals := sta.dv[len(sta.dv)-period:]
						for i := 0; i < period; i++ {
							if vals[i] < dvVal {
								lowNum += 1
							} else if vals[i] == dvVal {
								equalNum += 1
							}
						}
						if len(sta.dv) >= period*3 {
							sta.dv = vals
						}
						hitNum := lowNum + (equalNum+1)/2
						res.Append(hitNum * 100 / float64(period))
					} else {
						res.Append(math.NaN())
					}
				}
			}
		}
	}
	return res
}

/*
KDJ alias: stoch indicator;

period: 9, sm1: 3, sm2: 3

return (K, D, RSV)
*/
func KDJ(high *Series, low *Series, close *Series, period int, sm1 int, sm2 int) (*Series, *Series, *Series) {
	return KDJBy(high, low, close, period, sm1, sm2, "rma")
}

/*
KDJBy alias: talib stoch indicator;

period: 9, sm1: 3, sm2: 3

return (K, D, RSV)
*/
func KDJBy(high *Series, low *Series, close *Series, period int, sm1 int, sm2 int, maBy string) (*Series, *Series, *Series) {
	byVal, _ := kdjTypes[maBy]
	res := high.To("_kdj", period*100000+sm1*1000+sm2*10+byVal)
	if !res.Cached() {
		if !res.Cached() {
			rsv := Stoch(high, low, close, period)
			if maBy == "rma" {
				k := RMABy(rsv, sm1, 0, 50)
				d := RMABy(k, sm2, 0, 50)
				res.Append([]*Series{k, d, rsv})
			} else if maBy == "sma" {
				k := SMA(rsv, sm1)
				d := SMA(k, sm2)
				res.Append([]*Series{k, d, rsv})
			} else {
				panic(fmt.Sprintf("unknown maBy for KDJ: %s", maBy))
			}
		}
	}
	return res, res.Cols[0], res.Cols[1]
}

/*
PercentRank

calculates the percentile rank of a bar value in a data set.
*/
func PercentRank(obj *Series, period int) *Series {
	res := obj.To("_pecRk", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		inVal := obj.Get(0)
		if math.IsNaN(inVal) {
			res.Append(math.NaN())
		} else {
			vals := WrapFloatArr(res, period, inVal)
			if len(vals) < period {
				res.Append(math.NaN())
			} else {
				lowNum := float64(0)
				for i := 0; i < period-1; i++ {
					if vals[i] <= inVal {
						lowNum += 1
					}
				}
				res.Append(lowNum * 100 / float64(period))
			}
		}
	}
	return res
}

/*
RMI Relative Momentum Index
https://theforexgeek.com/relative-momentum-index/

period: 14, montLen: 3
*/
func RMI(obj *Series, period int, montLen int) *Series {
	res := obj.To("_rmi", period*1000+montLen)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		// These helper series are stateful. Include the complete RMI parameter
		// set in their cache key so calls that share montLen do not append to
		// the same series during one bar.
		maxChg := obj.To("_rmi_max_chg", period).To("_mont_len", montLen)
		minChg := obj.To("_rmi_min_chg", period).To("_mont_len", montLen)
		inVal := obj.Get(0)
		if math.IsNaN(inVal) {
			maxChg.Append(math.NaN())
			minChg.Append(math.NaN())
			res.Append(math.NaN())
		} else {
			arr := WrapFloatArr(res, montLen+1, inVal)
			if len(arr) < montLen+1 {
				maxChg.Append(math.NaN())
				minChg.Append(math.NaN())
				res.Append(math.NaN())
			} else {
				chgVal := inVal - arr[0]
				maxChg.Append(max(0, chgVal))
				minChg.Append(-min(0, chgVal))
				up := RMA(maxChg, period).Get(0)
				down := RMA(minChg, period).Get(0)
				var rmiVal = math.NaN()
				if down == 0 {
					rmiVal = 100
				} else if up == 0 {
					rmiVal = 0
				} else if !math.IsNaN(up) && !math.IsNaN(down) {
					rmiVal = 100 - (100 / (1 + up/down))
				}
				res.Append(rmiVal)
			}
		}
	}
	return res
}

/*
ROC rate of change

suggest period: 9
*/
func ROC(obj *Series, period int) *Series {
	res := obj.To("_roc", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		prevs, ok := res.More.([]float64)
		if !ok {
			res.DupMore = func(more interface{}) interface{} {
				return append([]float64{}, more.([]float64)...)
			}
		}
		curVal := obj.Get(0)
		if math.IsNaN(curVal) {
			res.Append(math.NaN())
		} else {
			prevs = append(prevs, curVal)
			if len(prevs) > period+1 {
				prevs = prevs[1:]
				res.More = prevs
			} else if len(prevs) <= period {
				res.More = prevs
				res.Append(math.NaN())
				return res
			}
			res.More = prevs
			preVal := prevs[0]
			var rocVal float64
			if preVal != 0 {
				rocVal = (curVal - preVal) / preVal * 100
			} else {
				rocVal = math.NaN() // 避免除以零
			}
			res.Append(rocVal)
		}
	}
	return res
}

/*
RSI Relative Strength Index 计算相对强度指数

suggest period: 14
*/
func RSI(obj *Series, period int) *Series {
	return rsiBy(obj, period, 0)
}

/*
STC colored indicator

period: 12  fast: 26  slow: 50  alpha: 0.5

https://www.tradingview.com/u/shayankm/
*/
func STC(obj *Series, period, fast, slow int, alpha float64) *Series {
	res := obj.To("_stc", pkey(ikey(period), ikey(fast), ikey(slow), fkey(alpha)))
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		s, _ := res.More.(*stcSta)
		if s == nil {
			s = &stcSta{
				prevDDD: math.NaN(),
				prevSTC: math.NaN(),
			}
			res.More = s
			res.DupMore = func(more interface{}) interface{} {
				m := more.(*stcSta)
				return &stcSta{append([]float64{}, m.macdHis...), append([]float64{}, m.dddHis...), m.prevDDD, m.prevSTC}
			}
		}
		// 1. 计算MACD差值
		fastEMA := EMA(obj, fast).Get(0)
		slowEMA := EMA(obj, slow).Get(0)
		macd := fastEMA - slowEMA
		if math.IsNaN(macd) {
			if !math.IsNaN(s.prevDDD) {
				s.macdHis = nil
				s.dddHis = nil
				s.prevDDD = math.NaN()
				s.prevSTC = math.NaN()
			}
			res.Append(math.NaN())
		} else {
			// 2. 维护MACD窗口（保持长度为Length）
			s.macdHis = append(s.macdHis, macd)
			if len(s.macdHis) > period {
				s.macdHis = s.macdHis[1:]
			}

			// 3. 计算第一层百分比
			ccccc := calcHLRangePct(s.macdHis, macd)

			// 4. 计算第一层平滑值
			ddd := ccccc
			if !math.IsNaN(s.prevDDD) {
				ddd = s.prevDDD + alpha*(ccccc-s.prevDDD)
			}
			s.prevDDD = ddd

			// 5. 维护DDD窗口
			s.dddHis = append(s.dddHis, ddd)
			if len(s.dddHis) > period {
				s.dddHis = s.dddHis[1:]
			}

			// 6. 计算第二层百分比
			dddddd := calcHLRangePct(s.dddHis, ddd)

			// 7. 计算最终STC值
			stc := dddddd
			if !math.IsNaN(s.prevSTC) {
				stc = s.prevSTC + alpha*(dddddd-s.prevSTC)
			}
			s.prevSTC = stc
			res.Append(stc)
		}
	}
	return res
}

/*
Stoch 100 * (close - lowest(low, period)) / (highest(high, period) - lowest(low, period))

use KDJ if you want to apply SMA/RMA to this

suggest period: 14
*/
func Stoch(high, low, close *Series, period int) *Series {
	res := high.To("_rsv", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		hhigh := Highest(high, period).Get(0)
		llow := Lowest(low, period).Get(0)
		maxChg := hhigh - llow
		if equalNearly(maxChg, 0) {
			res.Append(50.0)
		} else {
			res.Append((close.Get(0) - llow) / maxChg * 100)
		}
	}
	return res
}

/*
StochRSI StochasticRSI

rsiLen: 14, stochLen: 14, maK: 3, maD: 3

return [fastK, fastD]
*/
func StochRSI(obj *Series, rsiLen int, stochLen int, maK int, maD int) (*Series, *Series) {
	res := obj.To("_stoch_rsi", rsiLen*100000+stochLen*1000+maK*10+maD)
	if !res.Cached() {
		if !res.Cached() {
			rsi := RSI(obj, rsiLen)
			stochCol := Stoch(rsi, rsi, rsi, stochLen)
			smoothK := SMA(stochCol, maK)
			smoothD := SMA(smoothK, maD)
			res.Append([]float64{smoothK.Get(0), smoothD.Get(0)})
		}
	}
	return res, res.Cols[0]
}

/*
UpDown

vtype: 0 TradingView (Count consecutive times)

1 classic (abs count up to 1)
*/
func UpDown(obj *Series, vtype int) *Series {
	res := obj.To("_updn", vtype)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		old := res.Get(0)
		sub := obj.Get(0) - obj.Get(1)
		var resVal = math.NaN()
		if sub == 0 {
			resVal = 0
		} else if sub > 0 {
			if old > 0 && vtype == 0 {
				resVal = old + 1
			} else {
				resVal = 1
			}
		} else if sub < 0 {
			if old < 0 && vtype == 0 {
				resVal = old - 1
			} else {
				resVal = -1
			}
		} else if !math.IsNaN(obj.Get(0)) {
			resVal = 0
		}
		res.Append(resVal)
	}
	return res
}

/*
WillR William's Percent R

suggest period: 14
*/
func WillR(e *BarEnv, period int) *Series {
	res := e.Close.To("_williams_r", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		lowVal := Lowest(e.Low, period).Get(0)
		highVal := Highest(e.High, period).Get(0)
		rangeVal := highVal - lowVal
		if rangeVal == 0 {
			res.Append(math.NaN())
		} else {
			res.Append((e.Close.Get(0) - highVal) / rangeVal * 100)
		}
	}
	return res
}

// AO is the Awesome Oscillator: SMA(HL2, fast) - SMA(HL2, slow).
// The output is NaN until both simple moving averages are defined.
func AO(high, low *Series, fast, slow int) *Series {
	key := fast*100000 + slow
	res := high.To("_ao", key)
	if res.Cached() {
		return res
	}
	if fast < 1 || slow < 1 {
		res.Append(math.NaN())
		return res
	}
	price := HL2(high, low)
	res.Append(SMA(price, fast).Get(0) - SMA(price, slow).Get(0))
	return res
}

// DPO is the centered=False pandas-ta detrended price oscillator:
// close[t] - SMA(close, period)[t-shift], where shift=floor(period/2)+1.
// The stateful API emits NaN until both operands are defined.
func DPO(obj *Series, period int) *Series {
	res := obj.To("_dpo", period)
	if res.Cached() {
		return res
	}
	if period < 1 {
		res.Append(math.NaN())
		return res
	}
	type dpoState struct {
		values []float64
	}
	s, _ := res.More.(*dpoState)
	if s == nil {
		s = &dpoState{}
		res.More = s
		res.DupMore = func(v interface{}) interface{} {
			x := v.(*dpoState)
			return &dpoState{values: append([]float64(nil), x.values...)}
		}
	}
	cur := obj.Get(0)
	s.values = append(s.values, cur)
	shift := period/2 + 1
	if len(s.values) <= shift+period-1 {
		res.Append(math.NaN())
	} else {
		maEnd := len(s.values) - shift - 1
		old := s.values[maEnd]
		if math.IsNaN(old) {
			res.Append(math.NaN())
		} else {
			sum := 0.0
			valid := true
			for i := maEnd - period + 1; i <= maEnd; i++ {
				if math.IsNaN(s.values[i]) {
					valid = false
					break
				}
				sum += s.values[i]
			}
			if !valid {
				res.Append(math.NaN())
			} else {
				res.Append(cur - sum/float64(period))
			}
		}
	}
	return res
}

// Dpo preserves the historical mixed-case spelling.
func Dpo(obj *Series, period int) *Series { return DPO(obj, period) }

// MOM is the momentum (close - close[n]) indicator (TA-Lib MOM).
func MOM(obj *Series, period int) *Series {
	res := obj.To("_mom", period)
	if !res.Cached() {
		// Get is reverse-indexed and is the authoritative warm-up check;
		// Len can be capped by an environment's retention window.
		if period < 1 || math.IsNaN(obj.Get(period)) {
			res.Append(math.NaN())
		} else {
			res.Append(obj.Get(0) - obj.Get(period))
		}
	}
	return res
}

// RSI50 Relative Strength Index 计算相对强度指数-50
func RSI50(obj *Series, period int) *Series {
	return rsiBy(obj, period, 50)
}

// STOCH preserves the source-oriented call shape emitted by the converter:
// close/source, high, low, period.
func STOCH(close, high, low *Series, period int) *Series {
	return Stoch(high, low, close, period)
}

// StochF returns fast stochastic %K and %D (3-period SMA of %K).
func StochF(high, low, close *Series, period int, smooth ...int) (*Series, *Series) {
	d := 3
	if len(smooth) > 0 && smooth[0] > 0 {
		d = smooth[0]
	}
	key := period*1000 + d
	k := close.To("_stochf_k", key)
	raw := close.To("_stochf_raw", key)
	if k.Len() < close.Len() {
		value := math.NaN()
		if period > 0 && close.Len() >= period {
			hh, ll := high.Get(0), low.Get(0)
			for i := 1; i < period; i++ {
				hh = math.Max(hh, high.Get(i))
				ll = math.Min(ll, low.Get(i))
			}
			value = 50.0
			if hh != ll {
				value = (close.Get(0) - ll) / (hh - ll) * 100
			}
		}
		raw.Append(value)
		if close.Len() <= period+1 {
			k.Append(math.NaN())
		} else {
			k.Append(value)
		}
	}
	return k, SMA(raw, d)
}

// Ultimate Oscillator, with the canonical 7/14/28 periods and 4/2/1 weights.
func ULTOSC(high, low, close *Series, short, medium, long int) *Series {
	res := close.To("_ultosc", short*1000000+medium*1000+long)
	if !res.Cached() {
		prev := close.Back(1).Get(0)
		tr, bp := math.NaN(), math.NaN()
		if !math.IsNaN(prev) {
			tr = math.Max(high.Get(0), prev) - math.Min(low.Get(0), prev)
			bp = close.Get(0) - math.Min(low.Get(0), prev)
		}
		key := short*1000000 + medium*1000 + long
		bs := close.To("_ultbp", key)
		ts := close.To("_ulttr", key)
		if !bs.Cached() {
			bs.Append(bp)
			ts.Append(tr)
		}
		xs, xt := Sum(bs, short), Sum(ts, short)
		xm, tm := Sum(bs, medium), Sum(ts, medium)
		xl, tl := Sum(bs, long), Sum(ts, long)
		if xl.Len() == 0 || math.IsNaN(xl.Get(0)) || tl.Get(0) == 0 {
			res.Append(math.NaN())
		} else {
			res.Append(100 * (4*xs.Get(0)/xt.Get(0) + 2*xm.Get(0)/tm.Get(0) + xl.Get(0)/tl.Get(0)) / 7)
		}
	}
	return res
}

func Fisher(high, low *Series, period int) *Series {
	r := high.To("_fisher", period)
	if !r.Cached() {
		r.Append(last(tav.Fisher(hist(high), hist(low), period)))
	}
	return r
}

func STOCHF(close, high, low *Series, period int, smooth ...int) (*Series, *Series) {
	return StochF(high, low, close, period, smooth...)
}

func WilliamsPercent(high, low, close *Series, period int) *Series {
	r := close.To("_williams_percent", period)
	if !r.Cached() {
		r.Append(last(tav.WilliamsPercent(hist(high), hist(low), hist(close), period)))
	}
	return r
}

func rsiBy(obj *Series, period int, subVal float64) *Series {
	res := obj.To("_rsi", pkey(ikey(period), fkey(subVal)))
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		curVal := obj.Get(0)
		// 如果当前值为NaN，则跳过并返回NaN
		if math.IsNaN(curVal) {
			res.Append(math.NaN())
		} else {
			var more []float64
			if m, ok := res.More.([]float64); ok && len(m) == 4 {
				more = m
			} else {
				// 状态: [0:prevVal, 1:avgGain, 2:avgLoss, 3:validCount]
				more = []float64{math.NaN(), 0, 0, 0}
				res.More = more
				res.DupMore = func(more interface{}) interface{} {
					return append([]float64{}, more.([]float64)...)
				}
			}

			prevVal := more[0]
			valDelta := curVal - prevVal
			more[0] = curVal
			if math.IsNaN(prevVal) {
				res.Append(math.NaN())
			} else {
				// 从这里开始，我们有一个有效的delta可以计算
				validCount := more[3]
				validCount++
				more[3] = validCount

				var gainDelta, lossDelta float64
				if valDelta >= 0 {
					gainDelta = valDelta
				} else {
					lossDelta = -valDelta
				}
				if validCount > float64(period) {
					more[1] = (more[1]*float64(period-1) + gainDelta) / float64(period)
					more[2] = (more[2]*float64(period-1) + lossDelta) / float64(period)
				} else {
					more[1] += gainDelta / float64(period)
					more[2] += lossDelta / float64(period)
				}

				var resVal float64
				if validCount >= float64(period) {
					resVal = more[1]*100/(more[1]+more[2]) - subVal
				} else {
					resVal = math.NaN()
				}
				res.Append(resVal)
			}
		}
	}
	return res
}

/*
	CCI Commodity Channel Index

https://www.tradingview.com/support/solutions/43000502001-commodity-channel-index-cci/

suggest period: 20
*/
func CCI(obj *Series, args ...interface{}) *Series {
	period, source := 20, obj
	if len(args) == 1 {
		period, _ = args[0].(int)
	} else if len(args) == 3 {
		low, lowOK := args[0].(*Series)
		close, closeOK := args[1].(*Series)
		var periodOK bool
		period, periodOK = args[2].(int)
		if lowOK && closeOK && periodOK {
			source = HLC3(obj, low, close)
		}
	}
	res := obj.To("_cci", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		sma := SMA(source, period)
		meanDev := AvgDev(source, period)

		cciValue := (source.Get(0) - sma.Get(0)) / (0.015 * meanDev.Get(0))
		res.Append(cciValue)
	}
	return res
}

// ROCR is the TA-Lib rate-of-change ratio (current / value n bars ago).
func ROCR(obj *Series, period int) *Series {
	res := obj.To("_rocr", period)
	if !res.Cached() {
		prev := obj.Get(period)
		if period < 1 || math.IsNaN(prev) || prev == 0 {
			res.Append(math.NaN())
		} else {
			res.Append(obj.Get(0) / prev)
		}
	}
	return res
}

// TRIX is the one-bar percentage change of a triple-smoothed close.
func TRIX(obj *Series, period int) *Series {
	e := EMA(EMA(EMA(obj, period), period), period)
	return Change(e).Div(e.Back(1)).Mul(100)
}

// TSI is the double-smoothed true strength index (momentum / absolute momentum).
func TSI(obj *Series, short, long int) *Series {
	m := Change(obj)
	n := m.Abs()
	return EMA(EMA(m, short), long).Div(EMA(EMA(n, short), long)).Mul(100)
}
