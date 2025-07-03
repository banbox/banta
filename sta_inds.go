package banta

import (
	"fmt"
	"math"
	"slices"
)

/*
AvgPrice typical price=(h+l+c)/3
*/
func AvgPrice(e *BarEnv) *Series {
	res := e.Close.To("_avgp", 0)
	if res.Cached() {
		return res
	}
	avgPrice := (e.High.Get(0) + e.Low.Get(0) + e.Close.Get(0)) / 3
	return res.Append(avgPrice)
}

func HL2(h, l *Series) *Series {
	res := l.To("_hl", 0)
	if res.Cached() {
		return res
	}
	avgPrice := (h.Get(0) + l.Get(0)) / 2
	return res.Append(avgPrice)
}

// HLC3 typical price=(h+l+c)/3
func HLC3(h, l, c *Series) *Series {
	res := c.To("_hlc3", 0)
	if res.Cached() {
		return res
	}
	return res.Append((h.Get(0) + l.Get(0) + c.Get(0)) / 3)
}

type sumState struct {
	sumVal float64
	arr    []float64
}

func Sum(obj *Series, period int) *Series {
	res := obj.To("_sum", period)
	if res.Cached() {
		return res
	}
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
			return res.Append(sta.sumVal)
		}
	}
	return res.Append(math.NaN())
}

func SMA(obj *Series, period int) *Series {
	res := obj.To("_sma", period)
	if res.Cached() {
		return res
	}
	midObj := Sum(obj, period)
	if midObj.Len() >= period {
		return res.Append(midObj.Get(0) / float64(period))
	}
	return res.Append(math.NaN())
}

type moreVWMA struct {
	sumCost float64
	sumWei  float64
	costs   []float64
	volumes []float64
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
		return res.Append(math.NaN())
	} else {
		more.sumCost += cost
		more.sumWei += volVal
		more.volumes = append(more.volumes, volVal)
		more.costs = append(more.costs, cost)
	}
	if len(more.volumes) > period {
		more.sumCost -= more.costs[0]
		more.sumWei -= more.volumes[0]
		more.costs = more.costs[1:]
		more.volumes = more.volumes[1:]
	}
	if len(more.volumes) < period {
		return res.Append(math.NaN())
	}
	return res.Append(more.sumCost / more.sumWei)
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
	return res.Append(resVal)
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
	hash := period*1000 + initType*100
	if !math.IsNaN(initVal) {
		hash += int(initVal)
	}
	res := obj.To("_rma", hash)
	alpha := 1.0 / float64(period)
	return ewma(obj, res, period, alpha, initType, initVal)
}

type wmaSta struct {
	arr    []float64
	allSum float64
	weiSum float64
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
	val := obj.Get(0)
	if math.IsNaN(val) {
		return res.Append(math.NaN())
	}
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
		return res.Append(math.NaN())
	}
	sumWei := float64(period) * float64(period+1) * 0.5
	return res.Append(more.allSum / sumWei)
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
	half := WMA(obj, period/2).Get(0)
	wma := WMA(obj, period).Get(0)
	mid.Append(2*half - wma)
	return WMA(mid, maLen)
}

func TR(high *Series, low *Series, close *Series) *Series {
	res := high.To("_tr", 0)
	if res.Cached() {
		return res
	}
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
	return res.Append(resVal)
}

/*
ATR Average True Range

suggest period: 14
*/
func ATR(high *Series, low *Series, close *Series, period int) *Series {
	return RMA(TR(high, low, close), period)
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

func MACDBy(obj *Series, fast int, slow int, smooth int, initType int) (*Series, *Series) {
	res := obj.To("_macd", fast*1000+slow*100+smooth*10+initType)
	if !res.Cached() {
		short := EMABy(obj, fast, initType)
		longMA := EMABy(obj, slow, initType)
		macd := short.Sub(longMA)
		signal := EMABy(macd, smooth, initType)
		res.Append([]float64{macd.Get(0), signal.Get(0)})
	}
	return res, res.Cols[0]
}

func rsiBy(obj *Series, period int, subVal float64) *Series {
	res := obj.To("_rsi", period*100+int(subVal))
	if res.Cached() {
		return res
	}
	curVal := obj.Get(0)
	// 如果当前值为NaN，则跳过并返回NaN
	if math.IsNaN(curVal) {
		return res.Append(math.NaN())
	}

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
		return res.Append(math.NaN())
	}
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

	return res.Append(resVal)
}

/*
RSI Relative Strength Index 计算相对强度指数

suggest period: 14
*/
func RSI(obj *Series, period int) *Series {
	return rsiBy(obj, period, 0)
}

// RSI50 Relative Strength Index 计算相对强度指数-50
func RSI50(obj *Series, period int) *Series {
	return rsiBy(obj, period, 50)
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
	rsi := RSI(obj, period).Get(0)
	ud := RSI(UpDown(obj, vtype), upDn).Get(0)
	var rc float64
	if vtype == 0 {
		rc = PercentRank(ROC(obj, 1), roc).Get(0)
	} else {
		rc = ROC(obj, roc).Get(0)
	}
	return res.Append((rsi + ud + rc) / 3)
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
	return res.Append(resVal)
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
	inVal := obj.Get(0)
	if math.IsNaN(inVal) {
		return res.Append(math.NaN())
	}
	vals := WrapFloatArr(res, period, inVal)
	if len(vals) < period {
		return res.Append(math.NaN())
	}
	lowNum := float64(0)
	for i := 0; i < period-1; i++ {
		if vals[i] <= inVal {
			lowNum += 1
		}
	}
	return res.Append(lowNum * 100 / float64(period))
}

func Highest(obj *Series, period int) *Series {
	res := obj.To("_hh", period)
	if res.Cached() {
		return res
	}
	inVal := obj.Get(0)
	if math.IsNaN(inVal) {
		return res.Append(math.NaN())
	}
	// 获取周期内的数据
	values := WrapFloatArr(res, period, inVal)
	if len(values) < period {
		return res.Append(math.NaN())
	}
	return res.Append(slices.Max(values))
}

func HighestBar(obj *Series, period int) *Series {
	res := obj.To("_hhb", period)
	if res.Cached() {
		return res
	}
	if math.IsNaN(obj.Get(0)) {
		return res.Append(math.NaN())
	}
	values, ids := obj.RangeValid(0, period)
	if len(values) < period {
		return res.Append(math.NaN())
	}
	maxIdx, maxVal := -1, math.NaN()

	// 遍历以寻找非 NaN 的最大值
	for i, v := range values {
		// 如果 maxVal 是 NaN (说明这是第一个有效值) 或当前值更大，则更新
		if maxIdx < 0 || v > maxVal {
			maxVal = v
			maxIdx = ids[i]
		}
	}
	return res.Append(maxIdx)
}

func Lowest(obj *Series, period int) *Series {
	res := obj.To("_ll", period)
	if res.Cached() {
		return res
	}
	inVal := obj.Get(0)
	if math.IsNaN(inVal) {
		return res.Append(math.NaN())
	}
	// 获取周期内的数据
	values := WrapFloatArr(res, period, inVal)
	if len(values) < period {
		return res.Append(math.NaN())
	}
	return res.Append(slices.Min(values))
}

func LowestBar(obj *Series, period int) *Series {
	res := obj.To("_llb", period)
	if res.Cached() {
		return res
	}
	if math.IsNaN(obj.Get(0)) {
		return res.Append(math.NaN())
	}
	values, ids := obj.RangeValid(0, period)
	if len(values) < period {
		return res.Append(math.NaN())
	}
	minIdx, minVal := -1, math.NaN()

	// 遍历以寻找非 NaN 的最大值
	for i, v := range values {
		// 如果 maxVal 是 NaN (说明这是第一个有效值) 或当前值更大，则更新
		if minIdx < 0 || v < minVal {
			minVal = v
			minIdx = ids[i]
		}
	}
	return res.Append(minIdx)
}

/*
KDJ alias: stoch indicator;

period: 9, sm1: 3, sm2: 3

return (K, D, RSV)
*/
func KDJ(high *Series, low *Series, close *Series, period int, sm1 int, sm2 int) (*Series, *Series, *Series) {
	return KDJBy(high, low, close, period, sm1, sm2, "rma")
}

var (
	kdjTypes = map[string]int{
		"rma": 1,
		"sma": 2,
	}
)

/*
KDJBy alias: stoch indicator;

period: 9, sm1: 3, sm2: 3

return (K, D, RSV)
*/
func KDJBy(high *Series, low *Series, close *Series, period int, sm1 int, sm2 int, maBy string) (*Series, *Series, *Series) {
	byVal, _ := kdjTypes[maBy]
	res := high.To("_kdj", period*100000+sm1*1000+sm2*10+byVal)
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
	return res, res.Cols[0], res.Cols[1]
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
	hhigh := Highest(high, period).Get(0)
	llow := Lowest(low, period).Get(0)
	maxChg := hhigh - llow
	if equalNearly(maxChg, 0) {
		res.Append(50.0)
	} else {
		res.Append((close.Get(0) - llow) / maxChg * 100)
	}
	return res
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
		fac := -100 / float64(period)
		up := HighestBar(high, period+1).Mul(fac).Add(100)
		dn := LowestBar(low, period+1).Mul(fac).Add(100)
		osc := up.Sub(dn)
		res.Append([]*Series{up, osc, dn})
	}
	return res, res.Cols[0], res.Cols[1]
}

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
		meanVal := SMA(obj, period).Get(0)
		inVal := obj.Get(0)
		if math.IsNaN(inVal) {
			res.Append([]float64{math.NaN(), math.NaN()})
			return res, res.Cols[0]
		}
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
	return res, res.Cols[0]
}

func WrapFloatArr(res *Series, period int, inVal float64) []float64 {
	var more []float64
	if res.More == nil {
		more = make([]float64, 0, period)
		res.DupMore = func(more interface{}) interface{} {
			return append([]float64{}, more.([]float64)...)
		}
	} else {
		more = res.More.([]float64)
	}

	more = append(more, inVal)
	if len(more) < period {
		res.More = more
	} else {
		res.More = more[1:]
	}
	return more
}

/*
BBANDS Bollinger Bands 布林带指标

period: 20, stdUp: 2, stdDn: 2

return [upper, mid, lower]
*/
func BBANDS(obj *Series, period int, stdUp, stdDn float64) (*Series, *Series, *Series) {
	res := obj.To("_bb", period*10000+int(stdUp*1000)+int(stdDn*10))
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
	return res, res.Cols[0], res.Cols[1]
}

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
	inVal := obj.Get(0)
	if math.IsNaN(inVal) {
		return res.Append(math.NaN())
	}
	prevs := WrapFloatArr(res, 5, inVal)
	if len(prevs) < 5 {
		return res.Append(math.NaN())
	}
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
		return res.Append(resVal)
	}
	return res.Append(step)
}

/*
ADX Average Directional Index

suggest period: 14

return [maDX, plusDI, minusDI]
*/
func ADX(high *Series, low *Series, close *Series, period int) *Series {
	return ADXBy(high, low, close, period, 0, 0)
}

/*
	ADXBy Average Directional Index

method=0 classic ADX
method=1 TradingView "ADX and DI for v4"

suggest period: 14
smoothing: 0 to use period

return [maDX, plusDI, minusDI]
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

	plusDIVal := plusDI.Get(0)
	if math.IsNaN(plusDIVal) {
		dx.Append(math.NaN())
		adx.Append(math.NaN())
		return adx
	}
	minusDIVal := minusDI.Get(0)
	dx.Append(math.Abs(plusDIVal-minusDIVal) / (plusDIVal + minusDIVal) * 100)

	var maDX float64
	if method == 0 {
		maDX = RMA(dx, smoothing).Get(0)
	} else {
		maDX = SMA(dx, smoothing).Get(0)
	}
	adx.Append(maDX)
	return adx
}

type dmState struct {
	Num     int     // 计算次数
	DmPosMA float64 // 缓存DMPos的均值
	DmNegMA float64 // 缓存DMNeg的均值
	TRMA    float64 // 缓存TR的均值
}

/*
PluMinDI

suggest period: 14

return [plus di, minus di]
*/
func PluMinDI(high *Series, low *Series, close *Series, period int) (*Series, *Series) {
	return pluMinDIBy(high, low, close, period, 0)
}

func pluMinDIBy(high *Series, low *Series, close *Series, period, method int) (*Series, *Series) {
	plusDM, _ := pluMinDMBy(high, low, close, period, method)
	res := plusDM.To("_PluMinDI", period*10+method)
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

	return res, res.Cols[0]
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
		return res, res.Cols[0]
	} else {
		state.Num += 1
	}

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
	return res, res.Cols[0]
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
	prevs, ok := res.More.([]float64)
	if !ok {
		res.DupMore = func(more interface{}) interface{} {
			return append([]float64{}, more.([]float64)...)
		}
	}
	curVal := obj.Get(0)
	if math.IsNaN(curVal) {
		return res.Append(math.NaN())
	}
	prevs = append(prevs, curVal)
	if len(prevs) > period+1 {
		prevs = prevs[1:]
		res.More = prevs
	} else if len(prevs) <= period {
		res.More = prevs
		return res.Append(math.NaN())
	}
	res.More = prevs
	preVal := prevs[0]
	var rocVal float64
	if preVal != 0 {
		rocVal = (curVal - preVal) / preVal * 100
	} else {
		rocVal = math.NaN() // 避免除以零
	}
	return res.Append(rocVal)
}

// HeikinAshi return [open,high,low,close]
func HeikinAshi(e *BarEnv) (*Series, *Series, *Series, *Series) {
	res := e.Close.To("_heikin", 0)
	if !res.Cached() {
		ho := e.Open.To("_hka", 0)
		hh := e.High.To("_hka", 0)
		hl := e.Low.To("_hka", 0)
		hc := e.Close.To("_hka", 0)

		o, h, l, c := e.Open.Get(0), e.High.Get(0), e.Low.Get(0), e.Close.Get(0)

		po := ho.Get(0)
		if math.IsNaN(po) {
			ho.Append((o + c) / 2)
		} else {
			ho.Append((po + hc.Get(0)) / 2)
		}
		hcVal := (o + h + l + c) / 4
		hc.Append(hcVal)
		hoVal := ho.Get(0)
		hh.Append(max(h, hoVal, hcVal))
		hl.Append(min(l, hoVal, hcVal))

		res.Append([]*Series{ho, hh, hl, hc})
	}

	return res, res.Cols[0], res.Cols[1], res.Cols[2]
}

type tnrState struct {
	arr    []float64
	sumVal float64
	prevIn float64
	arrIn  []float64
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
		return res.Append(resVal)
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
	}
	return res.Append(resVal)
}

// AvgDev sum(abs(Vi - mean))/period
func AvgDev(obj *Series, period int) *Series {
	res := obj.To("_avgdev", period)
	if res.Cached() {
		return res
	}

	sma := SMA(obj, period)
	smaVal := sma.Get(0)
	inVal := obj.Get(0)
	if math.IsNaN(smaVal) || math.IsNaN(inVal) {
		return res.Append(math.NaN())
	}
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

	return res.Append(avgDev)
}

/*
	CCI Commodity Channel Index

https://www.tradingview.com/support/solutions/43000502001-commodity-channel-index-cci/

suggest period: 20
*/
func CCI(obj *Series, period int) *Series {
	res := obj.To("_cci", period)
	if res.Cached() {
		return res
	}
	sma := SMA(obj, period)
	meanDev := AvgDev(obj, period)

	cciValue := (obj.Get(0) - sma.Get(0)) / (0.015 * meanDev.Get(0))

	return res.Append(cciValue)
}

func moneyFlowVol(env *BarEnv) (float64, float64) {
	// Retrieve the latest values
	closeVal := env.Close.Get(0)
	high := env.High.Get(0)
	low := env.Low.Get(0)
	volume := env.Volume.Get(0)

	var multiplier float64

	// Money Flow Multiplier = [(Close - Low) - (High - Close)] / (High - Low)
	if high > low {
		multiplier = ((closeVal - low) - (high - closeVal)) / (high - low)
	}

	// Money Flow Volume = Money Flow Multiplier x Volume
	return multiplier, volume
}

type cmfState struct {
	mfSum    []float64
	volSum   []float64
	sumMfVal float64
	sumVol   float64
}

/*
CMF Chaikin Money Flow

https://www.tradingview.com/scripts/chaikinmoneyflow/?solution=43000501974

suggest period: 20
*/
func CMF(env *BarEnv, period int) *Series {
	res := env.Close.To("_cmf", period)
	if res.Cached() {
		return res
	}

	multiplier, volume := moneyFlowVol(env)
	mfVolume := multiplier * volume

	sta, _ := res.More.(*cmfState)
	if sta == nil {
		sta = &cmfState{}
		res.More = sta
		res.DupMore = func(more interface{}) interface{} {
			m := more.(*cmfState)
			return &cmfState{append([]float64{}, m.mfSum...), append([]float64{}, m.volSum...), m.sumMfVal, m.sumVol}
		}
	}

	var resVal = math.NaN()
	if math.IsNaN(mfVolume) || math.IsNaN(volume) || volume <= 0 {
	} else {
		// Sum the Money Flow Volumes and Volumes over the period
		sta.sumMfVal += mfVolume
		sta.sumVol += volume

		if len(sta.mfSum) < period {
			sta.mfSum = append(sta.mfSum, mfVolume)
			sta.volSum = append(sta.volSum, volume)
			if len(sta.mfSum) == period {
				resVal = sta.sumMfVal / sta.sumVol
			}
		} else {
			sta.sumMfVal -= sta.mfSum[0]
			sta.mfSum = append(sta.mfSum[1:], mfVolume)

			sta.sumVol -= sta.volSum[0]
			sta.volSum = append(sta.volSum[1:], volume)
			if sta.sumVol > 0 {
				// Calculate CMF = Sum(Money Flow Volume) / Sum(Volume)
				resVal = sta.sumMfVal / sta.sumVol
			}
		}
	}
	return res.Append(resVal)
}

// ADL Accumulation/Distribution Line
func ADL(env *BarEnv) *Series {
	adl := env.Close.To("_adl", 0)
	if adl.Cached() {
		return adl
	}
	multiplier, volume := moneyFlowVol(env)
	mfVolume := multiplier * volume

	adlValue := mfVolume
	if adl.Len() > 0 {
		adlValue += adl.Get(0)
	}
	return adl.Append(adlValue)
}

/*
ChaikinOsc Chaikin Oscillator

https://www.tradingview.com/support/solutions/43000501979-chaikin-oscillator/

short: 3, long: 10
*/
func ChaikinOsc(env *BarEnv, shortLen int, longLen int) *Series {
	res := env.Close.To("_chaikinosc", shortLen*1000+longLen)
	if res.Cached() {
		return res
	}
	adl := ADL(env)

	shortEma := EMA(adl, shortLen)
	longEma := EMA(adl, longLen)

	oscValue := shortEma.Get(0) - longEma.Get(0)
	return res.Append(oscValue)
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

	return res.Append(resVal)
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
	lowVal := Lowest(e.Low, period).Get(0)
	highVal := Highest(e.High, period).Get(0)
	rangeVal := highVal - lowVal
	if rangeVal == 0 {
		return res.Append(math.NaN())
	}
	return res.Append((e.Close.Get(0) - highVal) / rangeVal * 100)
}

/*
StochRSI StochasticRSI

rsiLen: 14, stochLen: 14, maK: 3, maD: 3

return [fastK, fastD]
*/
func StochRSI(obj *Series, rsiLen int, stochLen int, maK int, maD int) (*Series, *Series) {
	res := obj.To("_stoch_rsi", rsiLen*100000+stochLen*1000+maK*10+maD)
	if !res.Cached() {
		rsi := RSI(obj, rsiLen)
		stochCol := Stoch(rsi, rsi, rsi, stochLen)
		smoothK := SMA(stochCol, maK)
		smoothD := SMA(smoothK, maD)
		res.Append([]float64{smoothK.Get(0), smoothD.Get(0)})
	}
	return res, res.Cols[0]
}

type mfiState struct {
	posArr []float64
	negArr []float64
	sumPos float64
	sumNeg float64
	prev   float64
}

/*
MFI Money Flow Index

https://corporatefinanceinstitute.com/resources/career-map/sell-side/capital-markets/money-flow-index/

suggest period: 14
*/
func MFI(e *BarEnv, period int) *Series {
	res := e.Close.To("_mfi", period)
	if res.Cached() {
		return res
	}
	sta, _ := res.More.(*mfiState)
	if sta == nil {
		sta = &mfiState{sumNeg: math.NaN(), sumPos: math.NaN(), prev: math.NaN()}
		res.More = sta
		res.DupMore = func(more interface{}) interface{} {
			m := more.(*mfiState)
			return &mfiState{append([]float64{}, m.posArr...), append([]float64{}, m.negArr...), m.sumPos, m.sumNeg, m.prev}
		}
	}
	avgPrice := AvgPrice(e)
	price0 := avgPrice.Get(0)
	if math.IsNaN(price0) {
		return res.Append(math.NaN())
	}
	moneyFlow := price0 * e.Volume.Get(0)
	posFlow, negFlow := float64(0), float64(0)
	if price0 > sta.prev {
		posFlow = moneyFlow
	} else if price0 < sta.prev {
		negFlow = moneyFlow
	}
	sta.prev = price0
	if math.IsNaN(moneyFlow) {
		return res.Append(math.NaN())
	}
	var resVal = math.NaN()
	if math.IsNaN(sta.sumPos) || math.IsNaN(sta.sumNeg) {
		sta.posArr = []float64{posFlow}
		sta.negArr = []float64{negFlow}
		sta.sumPos = posFlow
		sta.sumNeg = negFlow
	} else {
		sta.sumPos += posFlow
		sta.sumNeg += negFlow
		sta.posArr = append(sta.posArr, posFlow)
		sta.negArr = append(sta.negArr, negFlow)
		if len(sta.posArr) >= period {
			if len(sta.posArr) > period {
				sta.sumPos -= sta.posArr[0]
				sta.sumNeg -= sta.negArr[0]
				sta.posArr = sta.posArr[1:]
				sta.negArr = sta.negArr[1:]
			}
			if sta.sumNeg > 0 {
				moneyFlowRatio := sta.sumPos / sta.sumNeg
				resVal = 100 - (100 / (1 + moneyFlowRatio))
			}
		}
	}
	return res.Append(resVal)
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
	maxChg := obj.To("_max_chg", montLen)
	minChg := obj.To("_min_chg", montLen)
	inVal := obj.Get(0)
	if math.IsNaN(inVal) {
		maxChg.Append(math.NaN())
		minChg.Append(math.NaN())
		return res.Append(math.NaN())
	}
	arr := WrapFloatArr(res, montLen+1, inVal)
	if len(arr) < montLen+1 {
		maxChg.Append(math.NaN())
		minChg.Append(math.NaN())
		return res.Append(math.NaN())
	}
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
	return res.Append(rmiVal)
}

func boolToHash(vals ...bool) int {
	result := 0
	for i, v := range vals {
		if v {
			result += 1 << i
		}
	}
	return result
}

/*
LinReg Linear Regression Moving Average

Linear Regression Moving Average (LINREG). This is a simplified version of a
Standard Linear Regression. LINREG is a rolling regression of one variable. A
Standard Linear Regression is between two or more variables.
*/
func LinReg(obj *Series, period int) *Series {
	return LinRegAdv(obj, period, false, false, false, false, false, false)
}

func LinRegAdv(obj *Series, period int, angle, intercept, degrees, r, slope, tsf bool) *Series {
	hash := period*100 + boolToHash(angle, intercept, degrees, r, slope, tsf)
	res := obj.To("_linreg", hash)
	if res.Cached() {
		return res
	}
	sumY := Sum(obj, period).Get(0)
	val := obj.Get(0)
	if math.IsNaN(val) {
		return res.Append(math.NaN())
	}
	arr := WrapFloatArr(res, period, val)
	if len(arr) < period || math.IsNaN(sumY) {
		return res.Append(math.NaN())
	}
	periodF := float64(period)
	var sumXY = float64(0)
	var sumX = periodF * float64(period+1) * 0.5
	sumY2 := float64(0)
	for i := 0; i < period; i++ {
		v := arr[i]
		sumXY += float64(i+1) * v
		if r {
			sumY2 += v * v
		}
	}
	sumX2 := sumX * (2*periodF + 1) / 3
	divisor := periodF*sumX2 - sumX*sumX
	m := (periodF*sumXY - sumX*sumY) / divisor
	if slope {
		return res.Append(m)
	}
	b := (sumY*sumX2 - sumX*sumXY) / divisor
	if intercept {
		return res.Append(b)
	}
	if angle {
		theta := math.Atan(m)
		if degrees {
			theta *= 180 / math.Pi
		}
		return res.Append(theta)
	}
	if r {
		rn := periodF*sumXY - sumX*sumY
		rd := math.Pow(divisor*(periodF*sumY2-sumY*sumY), 0.5)
		return res.Append(rn / rd)
	}
	if tsf {
		return res.Append(m*periodF + b)
	}
	return res.Append(m*(periodF-1) + b)
}

/*
CTI Correlation Trend Indicator

The Correlation Trend Indicator is an oscillator created by John Ehler in 2020.
It assigns a value depending on how close prices in that range are to following
a positively- or negatively-sloping straight line. Values range from -1 to 1.
This is a wrapper for LinRegAdv.

suggest period: 20
*/
func CTI(obj *Series, period int) *Series {
	return LinRegAdv(obj, period, false, false, false, true, false, false)
}

type cmdSta struct {
	subs   []float64
	sumPos float64
	sumNeg float64
	prevIn float64
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
	} else {
		return res.Append(math.NaN())
	}
	if len(sta.subs) < period {
		return res.Append(math.NaN())
	}
	return res.Append((sta.sumPos - sta.sumNeg) * 100 / (sta.sumPos + sta.sumNeg))
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
	atrSum := Sum(ATR(e.High, e.Low, e.Close, 1), period).Get(0)
	hh := Highest(e.High, period).Get(0)
	ll := Lowest(e.Low, period).Get(0)
	val := 100 * math.Log10(atrSum/(hh-ll)) / math.Log10(float64(period))
	return res.Append(val)
}

/*
ALMA Arnaud Legoux Moving Average

period:  window size. Default: 10

sigma: Smoothing value. Default 6.0

distOff: min 0 (smoother), max 1 (more responsive). Default: 0.85
*/
func ALMA(obj *Series, period int, sigma, distOff float64) *Series {
	res := obj.To("_alma", period*1000+int(sigma*100+distOff*100))
	if res.Cached() {
		return res
	}
	inVal := obj.Get(0)
	if math.IsNaN(inVal) {
		return res.Append(math.NaN())
	}
	arr := WrapFloatArr(res, period, inVal)
	if len(arr) < period {
		return res.Append(math.NaN())
	}
	m := distOff * (float64(period) - 1)
	s := float64(period) / sigma
	var windowSum, cumSum = float64(0), float64(0)
	for i := range arr {
		fi := float64(i)
		wei := math.Exp(-(fi - m) * (fi - m) / (2 * s * s))
		windowSum += wei * arr[period-i-1]
		cumSum += wei
	}
	return res.Append(windowSum / cumSum)
}

/*
Stiffness Indicator

maLen: 100, stiffLen: 60, stiffMa: 3
*/
func Stiffness(obj *Series, maLen, stiffLen, stiffMa int) *Series {
	bound := obj.To("_sti_bound", maLen)
	if !bound.Cached() {
		stdDev := StdDev(obj, maLen)
		bound.Append(SMA(obj, maLen).Get(0) - stdDev.Get(0)*0.2)
	}
	above := bound.To("_raw_gt", stiffLen)
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
	return EMA(Sum(above, stiffLen), stiffMa)
}

type dv2Sta struct {
	chl []float64
	dv  []float64
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
	// 如果当前输入值为 NaN，返回 NaN然后跳过，不重置状态
	if math.IsNaN(h0) || math.IsNaN(l0) || math.IsNaN(c0) {
		return res.Append(math.NaN())
	}
	// calc close*2/(high+low)
	den := (h0 + l0) / 2
	if den == 0 {
		return res.Append(math.NaN()) // 避免除以零，视为无效值
	}
	chl := c0/den - 1
	sta.chl = append(sta.chl, chl)
	// apply maLen to chl
	chlLen := len(sta.chl)
	if chlLen < maLen {
		sta.dv = append(sta.dv, 0)
		return res.Append(math.NaN())
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
			return res.Append(hitNum * 100 / float64(period))
		} else {
			return res.Append(math.NaN())
		}
	}
}

/*
UTBot UT Bot Alerts from TradingView
*/
func UTBot(c, atr *Series, rate float64) *Series {
	res := atr.To("_utBot", int(rate*10))
	if res.Cached() {
		return res
	}
	prevXATRTrailingStop, _ := res.More.(float64)
	nLoss := atr.Mul(rate).Get(0)
	// 计算动态止损线
	price := c.Get(0)
	prevSrc := c.Get(1)
	if math.IsNaN(nLoss) {
		return res.Append(math.NaN())
	}
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
		return res.Append(1)
	} else if price < xATRTrailingStop && below {
		return res.Append(-1)
	} else {
		return res.Append(0)
	}
}

// (cur-min)*100/(max-min)
func calcHLRangePct(his []float64, cur float64) float64 {
	if len(his) == 0 {
		return 0
	}
	// 查找窗口内的极值
	minVal, maxVal := math.MaxFloat64, -math.MaxFloat64
	for _, val := range his {
		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
	}

	// 计算百分比
	rangeSize := maxVal - minVal
	if rangeSize > 0 {
		return (cur - minVal) / rangeSize * 100
	}
	return 0
}

type stcSta struct {
	macdHis []float64 // MACD差值窗口
	dddHis  []float64 // DDD平滑值窗口
	prevDDD float64   // 前周期第一层平滑值
	prevSTC float64   // 前周期最终STC值
}

/*
STC colored indicator

period: 12  fast: 26  slow: 50  alpha: 0.5

https://www.tradingview.com/u/shayankm/
*/
func STC(obj *Series, period, fast, slow int, alpha float64) *Series {
	res := obj.To("_stc", fast*10000+slow*100+period)
	if res.Cached() {
		return res
	}
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
		return res.Append(math.NaN())
	}

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

	return res.Append(stc)
}
