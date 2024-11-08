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

func HL2(e *BarEnv) *Series {
	res := e.Close.To("_hl", 0)
	if res.Cached() {
		return res
	}
	avgPrice := (e.High.Get(0) + e.Low.Get(0)) / 2
	return res.Append(avgPrice)
}

type sumState struct {
	sumVal float64
	addLen int
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
	}
	curVal := obj.Get(0)
	if math.IsNaN(curVal) {
		// 输入值无效，重置，重新开始累计
		curVal = 0
		sta.sumVal = 0
		sta.addLen = 0
	} else {
		if sta.addLen < period {
			sta.sumVal += curVal
			sta.addLen += 1
		} else {
			oldVal := obj.Get(period)
			if math.IsNaN(oldVal) {
				sta.sumVal = 0
				sta.addLen = 0
			} else {
				sta.sumVal += curVal - oldVal
			}
		}
	}
	if sta.addLen < period {
		return res.Append(math.NaN())
	}
	return res.Append(sta.sumVal)
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
	len     int
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
	if more == nil || math.IsNaN(cost) {
		more = &moreVWMA{}
		res.More = more
	}
	if more.len >= period {
		oldVol := vol.Get(period)
		oldCost := price.Get(period) * oldVol
		more.sumCost -= oldCost
		more.sumWei -= oldVol
	}
	if !math.IsNaN(cost) {
		more.sumCost += cost
		more.sumWei += volVal
		more.len += 1
	}
	if more.len < period {
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
	inVal := obj.Get(0)
	var resVal float64
	if math.IsNaN(inVal) {
		resVal = inVal
	} else if res.Len() == 0 || math.IsNaN(res.Get(0)) {
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
		resVal = alpha*inVal + (1-alpha)*res.Get(0)
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
	arr, _ := res.More.([]float64)
	if math.IsNaN(val) {
		arr = nil
	} else if len(arr) >= period {
		arr = append(arr[1:], val)
	} else {
		arr = append(arr, val)
	}
	res.More = arr
	if len(arr) < period {
		return res.Append(math.NaN())
	}
	var sumVal = float64(0)
	var sumWei = float64(period) * float64(period+1) * 0.5
	for i := 0; i < period; i++ {
		sumVal += float64(i+1) * arr[i]
	}
	return res.Append(sumVal / sumWei)
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
	if high.Len() < 2 {
		return res.Append(math.NaN())
	} else {
		chigh, clow, pclose := high.Get(0), low.Get(0), close.Get(1)
		resVal := max(chigh-clow, math.Abs(chigh-pclose), math.Abs(clow-pclose))
		return res.Append(resVal)
	}
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
func MACD(obj *Series, fast int, slow int, smooth int) *Series {
	return MACDBy(obj, fast, slow, smooth, 0)
}

func MACDBy(obj *Series, fast int, slow int, smooth int, initType int) *Series {
	res := obj.To("_macd", fast*1000+slow*100+smooth*10+initType)
	if res.Cached() {
		return res
	}
	short := EMABy(obj, fast, initType)
	long := EMABy(obj, slow, initType)
	macd := short.Sub(long)
	signal := EMABy(macd, smooth, initType)
	return res.Append([]*Series{macd, signal})
}

func rsiBy(obj *Series, period int, subVal float64) *Series {
	res := obj.To("_rsi", period*100+int(subVal))
	if res.Cached() {
		return res
	}
	var more []float64
	if res.More == nil {
		more = []float64{math.NaN(), 0, 0}
		res.More = more
	} else {
		more = res.More.([]float64)
	}

	curVal := obj.Get(0)
	valDelta := curVal - more[0]
	more[0] = curVal
	if res.Len() == 0 || math.IsNaN(more[0]) {
		return res.Append(math.NaN())
	}

	var gainDelta, lossDelta float64
	if valDelta >= 0 {
		gainDelta, lossDelta = valDelta, 0
	} else {
		gainDelta, lossDelta = 0, -valDelta
	}
	if res.Len() > period {
		more[1] = (more[1]*float64(period-1) + gainDelta) / float64(period)
		more[2] = (more[2]*float64(period-1) + lossDelta) / float64(period)
	} else {
		more[1] += gainDelta / float64(period)
		more[2] += lossDelta / float64(period)
	}

	var resVal float64
	if res.Len() >= period {
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

func Highest(obj *Series, period int) *Series {
	res := obj.To("_hh", period)
	if res.Cached() {
		return res
	}
	if obj.Len() < period {
		return res.Append(math.NaN())
	}
	resVal := slices.Max(obj.Range(0, period))
	return res.Append(resVal)
}

func HighestBar(obj *Series, period int) *Series {
	res := obj.To("_hhb", period)
	if res.Cached() {
		return res
	}
	if obj.Len() < period {
		return res.Append(math.NaN())
	}
	data := obj.Range(0, period)
	maxIdx, maxVal := -1, 0.0
	for i, val := range data {
		if maxIdx < 0 || val > maxVal {
			maxVal = val
			maxIdx = i
		}
	}
	return res.Append(maxIdx)
}

func Lowest(obj *Series, period int) *Series {
	res := obj.To("_ll", period)
	if res.Cached() {
		return res
	}
	if obj.Len() < period {
		return res.Append(math.NaN())
	}
	resVal := slices.Min(obj.Range(0, period))
	return res.Append(resVal)
}

func LowestBar(obj *Series, period int) *Series {
	res := obj.To("_llb", period)
	if res.Cached() {
		return res
	}
	if obj.Len() < period {
		return res.Append(math.NaN())
	}
	data := obj.Range(0, period)
	minIdx, minVal := -1, 0.0
	for i, val := range data {
		if minIdx < 0 || val < minVal {
			minVal = val
			minIdx = i
		}
	}
	return res.Append(minIdx)
}

/*
KDJ alias: stoch indicator;

period: 9, sm1: 3, sm2: 3

return (K, D, RSV)
*/
func KDJ(high *Series, low *Series, close *Series, period int, sm1 int, sm2 int) *Series {
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
func KDJBy(high *Series, low *Series, close *Series, period int, sm1 int, sm2 int, maBy string) *Series {
	byVal, _ := kdjTypes[maBy]
	res := high.To("_kdj", period*100000+sm1*1000+sm2*10+byVal)
	if res.Cached() {
		return res
	}
	rsv := Stoch(high, low, close, period)
	if maBy == "rma" {
		k := RMABy(rsv, sm1, 0, 50)
		d := RMABy(k, sm2, 0, 50)
		return res.Append([]*Series{k, d, rsv})
	} else if maBy == "sma" {
		k := SMA(rsv, sm1)
		d := SMA(k, sm2)
		return res.Append([]*Series{k, d, rsv})
	} else {
		panic(fmt.Sprintf("unknown maBy for KDJ: %s", maBy))
	}
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
func Aroon(high *Series, low *Series, period int) *Series {
	res := high.To("_aroon", period)
	if res.Cached() {
		return res
	}
	fac := -100 / float64(period)
	up := HighestBar(high, period+1).Mul(fac).Add(100)
	dn := LowestBar(low, period+1).Mul(fac).Add(100)
	osc := up.Sub(dn)
	return res.Append([]*Series{up, osc, dn})
}

/*
	StdDev Standard Deviation 标准差和均值

suggest period: 20

return [stddev，sumVal]
*/
func StdDev(obj *Series, period int) *Series {
	return StdDevBy(obj, period, 0)
}

/*
	StdDevBy Standard Deviation

suggest period: 20

return [stddev，sumVal]
*/
func StdDevBy(obj *Series, period int, ddof int) *Series {
	res := obj.To("_sdev", period*10+ddof)
	if res.Cached() {
		return res
	}
	meanVal := SMA(obj, period).Get(0)
	var more []float64
	if res.More == nil {
		more = make([]float64, 0, period)
	} else {
		more = res.More.([]float64)
	}

	more = append(more, obj.Get(0))
	if len(more) < period {
		res.More = more
		return res.Append([]float64{math.NaN(), math.NaN()})
	}

	sumSqrt := 0.0
	for _, x := range more {
		sumSqrt += (x - meanVal) * (x - meanVal)
	}
	variance := sumSqrt / float64(period-ddof)
	stdDevVal := math.Sqrt(variance)
	res.More = more[1:]

	return res.Append([]float64{stdDevVal, meanVal})
}

/*
BBANDS Bollinger Bands 布林带指标

period: 20, stdUp: 2, stdDn: 2

return [upper, mid, lower]
*/
func BBANDS(obj *Series, period int, stdUp, stdDn float64) *Series {
	res := obj.To("_bb", period*10000+int(stdUp*1000)+int(stdDn*10))
	if res.Cached() {
		return res
	}
	stdDevCols := StdDev(obj, period).Cols
	dev, mean := stdDevCols[0].Get(0), stdDevCols[1].Get(0)
	if math.IsNaN(dev) {
		return res.Append([]float64{math.NaN(), math.NaN(), math.NaN()})
	}

	upper := mean + dev*stdUp
	lower := mean - dev*stdDn

	return res.Append([]float64{upper, mean, lower})
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
	sub4 := obj.Get(0) - obj.Get(4)
	if res.Len() == 0 || math.IsNaN(sub4) {
		return res.Append(math.NaN())
	}
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
	return ADXBy(high, low, close, period, 0)
}

type adxState struct {
	Num     int     // 计算次数
	DmPosMA float64 // 缓存DMPos的均值
	DmNegMA float64 // 缓存DMNeg的均值
	TRMA    float64 // 缓存TR的均值
}

/*
	ADXBy Average Directional Index

method=0 classic ADX
method=1 TradingView "ADX and DI for v4"

suggest period: 14

return [maDX, plusDI, minusDI]
*/
func ADXBy(high *Series, low *Series, close *Series, period int, method int) *Series {
	// 初始化相关的系列
	dx := close.To("_dx", period*1000+method)
	adx := close.To("_adx", period*1000+method)
	if adx.Cached() {
		return adx
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
	state, _ := adx.More.(*adxState)
	if state == nil {
		state = &adxState{}
		adx.More = state
	}
	state.Num += 1
	if math.IsNaN(tr) && math.IsNaN(close.Get(0)) {
		state.Num = 1
	}

	// calc Wilder's smoothing of DmH/DmL/TR
	alpha := 1 / float64(period)
	initLen := period
	if method == 1 {
		initLen = period + 1
	}
	if state.Num <= initLen {
		if math.IsNaN(tr) {
			state.DmPosMA = 0
			state.DmNegMA = 0
			state.TRMA = 0
		} else {
			state.DmPosMA += plusDM
			state.DmNegMA += minusDM
			state.TRMA += tr
		}
		if state.Num <= period {
			dx.Append(math.NaN())
			return adx.Append([]float64{math.NaN(), math.NaN(), math.NaN()})
		}
	} else {
		state.DmPosMA = state.DmPosMA*(1-alpha) + plusDM
		state.DmNegMA = state.DmNegMA*(1-alpha) + minusDM
		state.TRMA = state.TRMA*(1-alpha) + tr
	}

	// calc dx
	plusDI := 100 * state.DmPosMA / state.TRMA
	minusDI := 100 * state.DmNegMA / state.TRMA
	dx.Append(math.Abs(plusDI-minusDI) / (plusDI + minusDI) * 100)

	var maDX float64
	if method == 0 {
		maDX = RMA(dx, period).Get(0)
	} else {
		maDX = SMA(dx, period).Get(0)
	}
	return adx.Append([]float64{maDX, plusDI, minusDI})
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
	curVal := obj.Get(0)
	preVal := obj.Get(period)
	return res.Append((curVal - preVal) / preVal * 100)
}

// HeikinAshi 计算Heikin-Ashi
func HeikinAshi(e *BarEnv) *Series {
	res := e.Close.To("_heikin", 0)
	if res.Cached() {
		return res
	}

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

	return res.Append([]*Series{ho, hh, hl, hc})
}

type tnrState struct {
	arr    []float64
	sumVal float64
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
	curVal := math.Abs(obj.Get(0) - obj.Get(1))
	sta, _ := res.More.(*tnrState)
	if sta == nil {
		sta = &tnrState{}
		res.More = sta
	}
	var resVal = math.NaN()
	if math.IsNaN(curVal) {
		sta.arr = make([]float64, 0)
		sta.sumVal = 0
	} else {
		sta.sumVal += curVal
		if len(sta.arr) < period {
			sta.arr = append(sta.arr, curVal)
		} else {
			sta.sumVal -= sta.arr[0]
			sta.arr = append(sta.arr[1:], curVal)
			if sta.sumVal > 0 {
				diffVal := math.Abs(obj.Get(0) - obj.Get(period))
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
	sumDev := 0.0

	for i := 0; i < period; i++ {
		sumDev += math.Abs(obj.Get(i) - sma.Get(0))
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
	}

	var resVal = math.NaN()
	if math.IsNaN(mfVolume) || math.IsNaN(volume) || volume == 0 {
		sta.mfSum = make([]float64, 0)
		sta.volSum = make([]float64, 0)
		sta.sumMfVal = 0
		sta.sumVol = 0
	} else {
		// Sum the Money Flow Volumes and Volumes over the period
		sta.sumMfVal += mfVolume
		sta.sumVol += volume

		if len(sta.mfSum) < period {
			sta.mfSum = append(sta.mfSum, mfVolume)
			sta.volSum = append(sta.volSum, volume)
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
func ChaikinOsc(env *BarEnv, short int, long int) *Series {
	res := env.Close.To("_chaikinosc", short*1000+long)
	if res.Cached() {
		return res
	}
	adl := ADL(env)

	shortEma := EMA(adl, short)
	longEma := EMA(adl, long)

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

	effRatio := ER(obj, period).Get(0)
	var resVal = res.Get(0)

	if !math.IsNaN(effRatio) {
		fastV := 2 / float64(fast+1)
		slowV := 2 / float64(slow+1)
		alpha := math.Pow(effRatio*(fastV-slowV)+slowV, 2)
		curVal := obj.Get(0)
		if math.IsNaN(resVal) {
			resVal = curVal
		}
		resVal = alpha*curVal + (1-alpha)*resVal
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
	return res.Append((e.Close.Get(0) - highVal) / (highVal - lowVal) * 100)
}

/*
StochRSI StochasticRSI

rsiLen: 14, stochLen: 14, maK: 3, maD: 3
*/
func StochRSI(obj *Series, rsiLen int, stochLen int, maK int, maD int) *Series {
	res := obj.To("_stoch_rsi", rsiLen*100000+stochLen*1000+maK*10+maD)
	if res.Cached() {
		return res
	}
	rsi := RSI(obj, rsiLen)
	stochCol := Stoch(rsi, rsi, rsi, stochLen)
	smoothK := SMA(stochCol, maK)
	smoothD := SMA(smoothK, maD).Get(0)
	return res.Append([]float64{smoothK.Get(0), smoothD})
}

type mfiState struct {
	posArr []float64
	negArr []float64
	sumPos float64
	sumNeg float64
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
		sta = &mfiState{sumNeg: math.NaN(), sumPos: math.NaN()}
		res.More = sta
	}
	avgPrice := AvgPrice(e)
	price0, price1 := avgPrice.Get(0), avgPrice.Get(1)
	moneyFlow := price0 * e.Volume.Get(0)
	posFlow, negFlow := float64(0), float64(0)
	if price0 > price1 {
		posFlow = moneyFlow
	} else if price0 < price1 {
		negFlow = moneyFlow
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
		if len(sta.posArr) < period {
			sta.posArr = append(sta.posArr, posFlow)
			sta.negArr = append(sta.negArr, negFlow)
		} else {
			sta.sumPos -= sta.posArr[0]
			sta.sumNeg -= sta.negArr[0]
			sta.posArr = append(sta.posArr[1:], posFlow)
			sta.negArr = append(sta.negArr[1:], negFlow)
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
	chgVal := obj.Get(0) - obj.Get(montLen)
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
	arr, _ := res.More.([]float64)
	if math.IsNaN(val) {
		arr = nil
	} else if len(arr) >= period {
		arr = append(arr[1:], val)
	} else {
		arr = append(arr, val)
	}
	res.More = arr
	if len(arr) < period {
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
	} else {
		return res.Append(m*(periodF-1) + b)
	}
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
	val := obj.Get(0) - obj.Get(1)
	sta, _ := res.More.(*cmdSta)
	if sta == nil || math.IsNaN(val) {
		sta = &cmdSta{}
		res.More = sta
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
	if obj.Len() < period {
		return res.Append(math.NaN())
	}
	m := distOff * (float64(period) - 1)
	s := float64(period) / sigma
	var windowSum, cumSum = float64(0), float64(0)
	for i := 0; i < period; i++ {
		fi := float64(i)
		wei := math.Exp(-(fi - m) * (fi - m) / (2 * s * s))
		windowSum += wei * obj.Get(i)
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
		stdDev := StdDev(obj, maLen).Cols[0].Get(0)
		bound.Append(SMA(obj, maLen).Get(0) - stdDev*0.2)
	}
	above := bound.To("_raw_gt", stiffLen)
	if !above.Cached() {
		val := float64(0)
		if obj.Get(0) > bound.Get(0) {
			val = 100 / float64(stiffLen)
		}
		above.Append(val)
	}
	return EMA(Sum(above, stiffLen), stiffMa)
}
