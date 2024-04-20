package banta

import (
	"fmt"
	"math"
	"slices"
	"strconv"
)

func (e *BarEnv) GetSeries(key string) *Series {
	if ser, ok := e.Items[key]; ok {
		return ser
	}

	res := Series{e, nil, nil, key, 0, nil}
	e.Items[key] = &res
	return &res
}

func (e *BarEnv) OnBar(barMs int64, open, high, low, close, volume float64) {
	if e.TimeStop > barMs {
		panic(fmt.Errorf("%s/%s old Bar Receive: %d, Current: %d", e.Symbol, e.TimeFrame, barMs, e.TimeStop))
	}
	e.TimeStart = barMs
	e.TimeStop = barMs + e.TFMSecs
	e.BarNum += 1
	if e.Open == nil {
		e.Open = &Series{e, []float64{open}, nil, "o", barMs, nil}
		e.High = &Series{e, []float64{high}, nil, "h", barMs, nil}
		e.Low = &Series{e, []float64{low}, nil, "l", barMs, nil}
		e.Close = &Series{e, []float64{close}, nil, "c", barMs, nil}
		e.Volume = &Series{e, []float64{volume}, nil, "v", barMs, nil}
		e.Items = map[string]*Series{
			"o": e.Open,
			"h": e.High,
			"l": e.Low,
			"c": e.Close,
			"v": e.Volume,
		}
		if e.MaxCache == 0 {
			// 默认保留1000个
			e.MaxCache = 1000
		}
	} else {
		e.Open.Time = barMs
		e.Open.Data = append(e.Open.Data, open)
		e.High.Time = barMs
		e.High.Data = append(e.High.Data, high)
		e.Low.Time = barMs
		e.Low.Data = append(e.Low.Data, low)
		e.Close.Time = barMs
		e.Close.Data = append(e.Close.Data, close)
		e.Volume.Time = barMs
		e.Volume.Data = append(e.Volume.Data, volume)
		e.TrimOverflow()
	}
}

func (e *BarEnv) Reset() {
	e.TimeStart = 0
	e.TimeStop = 0
	e.BarNum = 0
	e.Open = nil
	e.High = nil
	e.Low = nil
	e.Close = nil
	e.Volume = nil
	e.Items = nil
	e.XLogs = nil
}

func (e *BarEnv) TrimOverflow() {
	dataLen := e.Close.Len()
	trimLen := int(float64(e.MaxCache) * 1.5)
	if dataLen < trimLen || trimLen <= 0 {
		return
	}
	e.Open.Cut(e.MaxCache)
	e.High.Cut(e.MaxCache)
	e.Low.Cut(e.MaxCache)
	e.Close.Cut(e.MaxCache)
	if e.Items != nil {
		for _, se := range e.Items {
			se.Cut(e.MaxCache)
		}
	}
}

func (s *Series) Set(obj interface{}) *Series {
	if !s.Cached() {
		return s.Append(obj)
	}
	return s
}

func (s *Series) Append(obj interface{}) *Series {
	if s.Time >= s.Env.TimeStop {
		panic(ErrRepeatAppend)
	}
	s.Time = s.Env.TimeStop
	if val, ok := obj.(float64); ok {
		s.Data = append(s.Data, val)
	} else if val, ok := obj.(int); ok {
		s.Data = append(s.Data, float64(val))
	} else if arr, ok := obj.([]float64); ok {
		for i, v := range arr {
			if i >= len(s.Cols) {
				key := fmt.Sprintf("%s_%d", s.Key, i)
				col := s.Env.GetSeries(key)
				s.Cols = append(s.Cols, col)
				col.Append(v)
			} else {
				col := s.Cols[i]
				col.Append(v)
			}
		}
	} else if cols, ok := obj.([]*Series); ok {
		s.Cols = cols
	} else {
		fmt.Printf("invalid val for Series.Append: %t", obj)
		panic(ErrInvalidSeriesVal)
	}
	return s
}

func (s *Series) Cached() bool {
	return s.Time >= s.Env.TimeStop
}

func (s *Series) Get(i int) float64 {
	if len(s.Cols) > 0 {
		panic(fmt.Errorf("Get Val on Merged Series!"))
	}
	allLen := len(s.Data)
	if i < 0 || i >= allLen {
		return math.NaN()
	}
	return s.Data[allLen-i-1]
}

/*
Range 获取范围内的值。
start 起始位置，0是最近的
stop 结束位置，不含
*/
func (s *Series) Range(start, stop int) []float64 {
	allLen := len(s.Data)
	_start := max(allLen-stop, 0)
	_stop := min(allLen-start, allLen)
	if _start >= _stop {
		return []float64{}
	}
	res := s.Data[_start:_stop]
	tmp := make([]float64, len(res))
	copy(tmp, res)
	slices.Reverse(tmp)
	return tmp
}

func (s *Series) Add(obj interface{}) *Series {
	return s.apply(obj, "%s+%s", false, func(a, b float64) float64 {
		return a + b
	})
}

func (s *Series) Sub(obj interface{}) *Series {
	return s.apply(obj, "%s-%s", false, func(a, b float64) float64 {
		return a - b
	})
}

func (s *Series) Mul(obj interface{}) *Series {
	return s.apply(obj, "%s*%s", false, func(a, b float64) float64 {
		return a * b
	})
}

func (s *Series) Abs() *Series {
	if len(s.Cols) > 0 {
		panic(ErrGetDataOfMerged)
	}
	res := s.Env.GetSeries(fmt.Sprintf("abs(%s)", s.Key))
	res.Append(math.Abs(s.Get(0)))
	return res
}

func (s *Series) Len() int {
	if len(s.Cols) > 0 {
		return s.Cols[0].Len()
	}
	return len(s.Data)
}

func (s *Series) Cut(keepNum int) {
	if len(s.Cols) > 0 {
		for _, col := range s.Cols {
			col.Cut(keepNum)
		}
		return
	}
	curLen := len(s.Data)
	if curLen <= keepNum {
		return
	}
	s.Data = s.Data[curLen-keepNum:]
}

func (s *Series) Back(num int) *Series {
	res := s.Env.GetSeries(fmt.Sprintf("%s_mv%v", s.Key, num))
	if !res.Cached() {
		endPos := len(s.Data) - num
		if endPos > 0 {
			res.Data = s.Data[:endPos]
		} else {
			res.Data = nil
		}
		res.Time = s.Env.TimeStop
	}
	return res
}

func (s *Series) apply(obj interface{}, text string, isRev bool, calc func(float64, float64) float64) *Series {
	if len(s.Cols) > 0 {
		panic(ErrGetDataOfMerged)
	}
	key2, val2 := keyVal(obj)
	var key1 = s.Key
	var val1 = s.Get(0)
	if isRev {
		key1, key2 = key2, key1
		val1, val2 = val2, val1
	}
	res := s.Env.GetSeries(fmt.Sprintf(text, key1, key2))
	res.Append(calc(val1, val2))
	return res
}

func keyVal(obj interface{}) (string, float64) {
	if ser, ok := obj.(*Series); ok {
		return ser.Key, ser.Get(0)
	} else if intVal, ok := obj.(int); ok {
		return strconv.Itoa(intVal), float64(intVal)
	} else if flt32Val, ok := obj.(float32); ok {
		return strconv.FormatFloat(float64(flt32Val), 'f', -1, 64), float64(flt32Val)
	} else if fltVal, ok := obj.(float64); ok {
		return strconv.FormatFloat(fltVal, 'f', -1, 64), fltVal
	} else {
		fmt.Printf("invalid val for Series.keyVal: %t", obj)
		panic(ErrInvalidSeriesVal)
	}
}

/*
Cross 计算最近一次交叉的距离。如果两个都变化，则两个都必须是序列。或者一个是常数一个是序列
返回值：正数上穿，负数下穿，0表示未知或重合；abs(ret) - 1表示交叉点与当前bar的距离
*/
func Cross(obj1 interface{}, obj2 interface{}) int {
	var env *BarEnv
	if ser1, ok1 := obj1.(*Series); ok1 {
		env = ser1.Env
	} else if ser2, ok2 := obj2.(*Series); ok2 {
		env = ser2.Env
	} else {
		panic(fmt.Errorf("one of obj1 or obj2 must be Series, %t %t", obj1, obj2))
	}
	k1, v1 := keyVal(obj1)
	k2, v2 := keyVal(obj2)
	xKey := fmt.Sprintf("%s_xup_%s", k1, k2)
	var log *CrossLog
	if val, ok := env.XLogs[xKey]; ok {
		log = val
	} else {
		log = &CrossLog{xKey, math.NaN(), []*XState{}}
		if env.XLogs == nil {
			env.XLogs = map[string]*CrossLog{}
		}
		env.XLogs[xKey] = log
	}
	diffVal := v1 - v2
	if diffVal != 0 {
		if math.IsNaN(log.PrevVal) {
			log.PrevVal = diffVal
		} else {
			if factor := log.PrevVal * diffVal; factor < 0 {
				log.PrevVal = diffVal
				log.Hist = append(log.Hist, &XState{numSign(diffVal), env.BarNum})
			}
		}
	}
	if len(log.Hist) > 0 {
		state := log.Hist[len(log.Hist)-1]
		return state.Sign * (env.BarNum - state.BarNum + 1)
	}
	return 0
}
