package banta

import (
	"fmt"
	"math"
	"slices"
)

/*
AvgPrice 平均价格=(h+l+c)/3
*/
func AvgPrice(e *BarEnv) *Series {
	res := e.Close.To("_avgp", 0)
	if res.Cached() {
		return res
	}
	avgPrice := (e.High.Get(0) + e.Low.Get(0) + e.Close.Get(0)) / 3
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
VWMA 成交量加权平均价格
公式：sum(price*volume)/sum(volume)
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
EMA 指数移动均线 最近一个权重：2/(n+1)
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
RMA 相对移动均线

	和EMA区别是：分子分母都减一
	最近一个权重：1/n
*/
func RMA(obj *Series, period int) *Series {
	return RMABy(obj, period, 0, math.NaN())
}

/*
RMABy 相对移动均线

	和EMA区别是：分子分母都减一
	最近一个权重：1/n

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

func ATR(high *Series, low *Series, close *Series, period int) *Series {
	return RMA(TR(high, low, close), period)
}

/*
MACD 计算MACD指标。
国外主流使用init_type=0，MyTT和国内主要使用init_type=1
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

func rsi(obj *Series, period int, subVal float64) *Series {
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

// RSI 计算相对强度指数
func RSI(obj *Series, period int) *Series {
	return rsi(obj, period, 0)
}

// RSI50 计算相对强度指数-50
func RSI50(obj *Series, period int) *Series {
	return rsi(obj, period, 50)
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
KDJ 也称为：Stoch随机指标。返回k, d
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

func KDJBy(high *Series, low *Series, close *Series, period int, sm1 int, sm2 int, maBy string) *Series {
	byVal, _ := kdjTypes[maBy]
	res := high.To("_kdj", period*100000+sm1*1000+sm2*10+byVal)
	if res.Cached() {
		return res
	}
	rsv := high.To("_rsv", period)
	if !rsv.Cached() {
		hhigh := Highest(high, period).Get(0)
		llow := Lowest(low, period).Get(0)
		maxChg := hhigh - llow
		if equalNearly(maxChg, 0) {
			rsv.Append(50.0)
		} else {
			rsv.Append((close.Get(0) - llow) / maxChg * 100)
		}
	}
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
Aroon 阿隆指标  反映了一段时间内出现最高价和最低价距离当前时间的远近。
AroonUp: (period - HighestBar(high, period+1)) / period * 100
AroonDn: (period - LowestBar(low, period+1)) / period * 100
Osc: AroonUp - AroonDn
返回：AroonUp, Osc, AroonDn
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
	StdDev 计算标准差和均值

返回：stddev，mean
*/
func StdDev(obj *Series, period int) *Series {
	return StdDevBy(obj, period, 0)
}

/*
	StdDevBy 计算标准差和均值

返回：stddev，mean
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

// BBANDS 布林带指标。返回：upper, mean, lower
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
	TD 计算Tom DeMark Sequence（狄马克序列）

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

type AdxState struct {
	Num    int     // 计算次数
	DmHSum float64 // 缓存初始DmH的和
	DmLSum float64 // 缓存初始DmL的和
	DmHMA  float64 // 缓存DMH的均值
	DmLMA  float64 // 缓存DML的均值
	TRMA   float64 // 缓存TR的均值
}

/*
	ADX 计算平均趋向指标

参考TradingView的社区ADX指标。与tdlib略有不同
*/
func ADX(high *Series, low *Series, close *Series, period int) *Series {
	// 初始化相关的系列
	dx := close.To("_dx", period)
	adx := close.To("_adx", period)

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
	tr := TR(high, low, close)
	var state *AdxState
	if adx.More == nil {
		state = &AdxState{}
		adx.More = state
	} else {
		state = adx.More.(*AdxState)
	}
	state.Num += 1

	if state.Num <= period+1 {
		state.DmHSum += plusDM
		state.DmLSum += minusDM
		if tr.Len() > 1 {
			state.TRMA += tr.Get(0)
		}
		if state.Num <= period {
			dx.Append(math.NaN())
			return adx.Append([]float64{math.NaN(), math.NaN(), math.NaN()})
		}
		state.DmHMA = state.DmHSum
		state.DmLMA = state.DmLSum
	} else {
		state.DmHMA = state.DmHMA*(1-1/float64(period)) + plusDM
		state.DmLMA = state.DmLMA*(1-1/float64(period)) + minusDM
		state.TRMA = state.TRMA*(1-1/float64(period)) + tr.Get(0)
	}

	plusDI := 100 * state.DmHMA / state.TRMA
	minusDI := 100 * state.DmLMA / state.TRMA
	dx.Append(math.Abs(plusDI-minusDI) / (plusDI + minusDI) * 100)

	smaDX := SMA(dx, period).Get(0)
	return adx.Append([]float64{smaDX, plusDI, minusDI})
}

/*
ROC rate of change
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
