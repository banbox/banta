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

// RSI 计算相对强度指数
func RSI(obj *Series, period int) *Series {
	return rsiBy(obj, period, 0)
}

// RSI50 计算相对强度指数-50
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

返回：stddev，sumVal
*/
func StdDev(obj *Series, period int) *Series {
	return StdDevBy(obj, period, 0)
}

/*
	StdDevBy 计算标准差和均值

返回：stddev，sumVal
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

// BBANDS 布林带指标。返回：upper, sumVal, lower
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

/*
ADX Average Directional Index
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

// TNR Trend to Noise Ratio / Efficiency Ratio
func TNR(obj *Series, period int) *Series {
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
ChaikinOsc
https://www.tradingview.com/support/solutions/43000501979-chaikin-oscillator/
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

// KAMA Kaufman Adaptive Moving Average
func KAMA(obj *Series, period int) *Series {
	return KAMABy(obj, period, 2, 30)
}

// KAMABy Kaufman Adaptive Moving Average
func KAMABy(obj *Series, period int, fast, slow int) *Series {
	res := obj.To("_kama", period*10000+slow*100+fast)
	if res.Cached() {
		return res
	}

	effRatio := TNR(obj, period).Get(0)
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

// Stoch 100 * (close - lowest(low, period)) / (highest(high, period) - lowest(low, period))
func Stoch(obj, high, low *Series, period int) *Series {
	res := high.To("_stoch", period)
	if res.Cached() {
		return res
	}
	lowVal := Lowest(low, period).Get(0)
	highVal := Highest(high, period).Get(0)
	return res.Append((obj.Get(0) - lowVal) / (highVal - lowVal) * 100)
}

func WillR(e *BarEnv, period int) *Series {
	res := e.Close.To("_williams_r", period)
	if res.Cached() {
		return res
	}
	lowVal := Lowest(e.Low, period).Get(0)
	highVal := Highest(e.High, period).Get(0)
	return res.Append((e.Close.Get(0) - highVal) / (highVal - lowVal) * 100)
}

// StochRSI StochasticRSI
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
