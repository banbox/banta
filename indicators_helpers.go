package banta

import (
	"math"
	"math/bits"
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

// pkey hashes raw integer/float bits without interface boxing. Use it for
// keys that include full-precision float parameters; integer-only hot paths
// use direct arithmetic keys.
func pkey(values ...uint64) int {
	switch len(values) {
	case 1:
		return int(values[0])
	case 2:
		if values[0] <= 0xffffffff && values[1] <= 0xffffffff {
			return int(values[0]<<32 | values[1])
		}
	case 3:
		if values[0] <= 0xffff && values[1] <= 0xffff && values[2] <= 0xffff {
			return int(values[0]<<32 | values[1]<<16 | values[2])
		}
	case 4:
		if values[0] <= 0xffff && values[1] <= 0xffff && values[2] <= 0xffff && values[3] <= 0xffff {
			return int(values[0]<<48 | values[1]<<32 | values[2]<<16 | values[3])
		}
	}
	if len(values) == 2 {
		return int(values[0] ^ bits.RotateLeft64(values[1], 23))
	}
	if len(values) == 3 {
		return int(values[0] ^ bits.RotateLeft64(values[1], 23) ^ bits.RotateLeft64(values[2], 47))
	}
	if len(values) == 4 {
		return int(values[0] ^ bits.RotateLeft64(values[1], 23) ^ bits.RotateLeft64(values[2], 47) ^ bits.RotateLeft64(values[3], 11))
	}
	h := uint64(len(values)) + 0x9e3779b97f4a7c15
	for _, value := range values {
		h ^= value + 0x9e3779b97f4a7c15 + (h << 6) + (h >> 2)
	}
	h ^= h >> 30
	h *= 0xbf58476d1ce4e5b9
	h ^= h >> 27
	h *= 0x94d049bb133111eb
	h ^= h >> 31
	return int(h)
}

func ikey(value int) uint64     { return uint64(int64(value)) }
func fkey(value float64) uint64 { return math.Float64bits(value) }

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
