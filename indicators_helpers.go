package banta

import (
	"math"
)

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

	if len(more) < period {
		more = append(more, inVal)
	} else {
		copy(more, more[1:])
		more[len(more)-1] = inVal
	}
	res.More = more
	return more
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

func hist(s *Series) []float64 {
	n := s.Len()
	r := make([]float64, n)
	for i := 0; i < n; i++ {
		r[n-1-i] = s.Get(i)
	}
	return r
}

func last(v []float64) float64 {
	if len(v) == 0 {
		return math.NaN()
	}
	return v[len(v)-1]
}

func pkey(p int, m float64) int { return p*100000 + int(m*1000) }

type cmdSta struct {
	subs   []float64
	sumPos float64
	sumNeg float64
	prevIn float64
}

type cmfState struct {
	mfSum    []float64
	volSum   []float64
	sumMfVal float64
	sumVol   float64
}

type dmState struct {
	Num     int     // 计算次数
	DmPosMA float64 // 缓存DMPos的均值
	DmNegMA float64 // 缓存DMNeg的均值
	TRMA    float64 // 缓存TR的均值
}

type dv2Sta struct {
	chl []float64
	dv  []float64
}

type mfiState struct {
	posArr []float64
	negArr []float64
	sumPos float64
	sumNeg float64
	prev   float64
}

type moreVWMA struct {
	sumCost float64
	sumWei  float64
	costs   []float64
	volumes []float64
}

type stcSta struct {
	macdHis []float64 // MACD差值窗口
	dddHis  []float64 // DDD平滑值窗口
	prevDDD float64   // 前周期第一层平滑值
	prevSTC float64   // 前周期最终STC值
}

type sumState struct {
	sumVal float64
	arr    []float64
}

type tnrState struct {
	arr    []float64
	sumVal float64
	prevIn float64
	arrIn  []float64
}

type vwapState struct {
	sumCost   float64
	sumVolume float64
}

type wmaSta struct {
	arr    []float64
	allSum float64
	weiSum float64
}

var (
	kdjTypes = map[string]int{
		"rma": 1,
		"sma": 2,
	}
)
