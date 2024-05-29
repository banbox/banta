package banta

import (
	"fmt"
	"math"
	"slices"
)

func (e *BarEnv) OnBar(barMs int64, open, high, low, close, volume, info float64) {
	if e.TimeStop > barMs {
		panic(fmt.Errorf("%s/%s old Bar Receive: %d, Current: %d", e.Symbol, e.TimeFrame, barMs, e.TimeStop))
	}
	e.TimeStart = barMs
	e.TimeStop = barMs + e.TFMSecs
	e.BarNum += 1
	if e.Open == nil {
		e.Open = e.NewSeries([]float64{open})
		e.High = e.NewSeries([]float64{high})
		e.Low = e.NewSeries([]float64{low})
		e.Close = e.NewSeries([]float64{close})
		e.Volume = e.NewSeries([]float64{volume})
		e.Info = e.NewSeries([]float64{info})
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
		e.Info.Time = barMs
		e.Info.Data = append(e.Info.Data, info)
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
	e.Info = nil
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
}

func (e *BarEnv) NewSeries(data []float64) *Series {
	subs := make(map[string]map[int]*Series)
	xlogs := make(map[int]*CrossLog)
	res := &Series{e.VNum, e, data, nil, e.TimeStart, nil, subs, xlogs}
	e.VNum += 1
	return res
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
				col := s.To("_", i)
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
	if len(s.Cols) > 0 {
		panic(ErrGetDataOfMerged)
	}
	res, val := s.objVal("_add", obj)
	return res.Append(s.Get(0) + val)
}

func (s *Series) Sub(obj interface{}) *Series {
	if len(s.Cols) > 0 {
		panic(ErrGetDataOfMerged)
	}
	res, val := s.objVal("_sub", obj)
	return res.Append(s.Get(0) - val)
}

func (s *Series) Mul(obj interface{}) *Series {
	if len(s.Cols) > 0 {
		panic(ErrGetDataOfMerged)
	}
	res, val := s.objVal("_mul", obj)
	return res.Append(s.Get(0) * val)
}

func (s *Series) Abs() *Series {
	if len(s.Cols) > 0 {
		panic(ErrGetDataOfMerged)
	}
	res := s.To("_abs", 0)
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
	for _, dv := range s.Subs {
		for _, v := range dv {
			v.Cut(keepNum)
		}
	}
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
	res := s.To("_back", num)
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

func (s *Series) objVal(rel string, obj interface{}) (*Series, float64) {
	if ser, ok := obj.(*Series); ok {
		return s.To(rel, ser.ID), ser.Get(0)
	} else if intVal, ok := obj.(int); ok {
		return s.To(rel, intVal), float64(intVal)
	} else if flt32Val, ok := obj.(float32); ok {
		return s.To(rel, int(flt32Val*10)), float64(flt32Val)
	} else if fltVal, ok := obj.(float64); ok {
		return s.To(rel, int(fltVal*10)), fltVal
	} else {
		fmt.Printf("invalid val for Series.objVal: %t", obj)
		panic(ErrInvalidSeriesVal)
	}
}

func (s *Series) To(k string, v int) *Series {
	sub, _ := s.Subs[k]
	if sub == nil {
		sub = make(map[int]*Series)
		s.Subs[k] = sub
	}
	old, _ := sub[v]
	if old == nil {
		old = s.Env.NewSeries(nil)
		sub[v] = old
	}
	return old
}

/*
Cross 计算最近一次交叉的距离。如果两个都变化，则两个都必须是序列。或者一个是常数一个是序列
返回值：正数上穿，负数下穿，0表示未知或重合；abs(ret) - 1表示交叉点与当前bar的距离
*/
func Cross(se *Series, obj2 interface{}) int {
	var env = se.Env
	var key int
	var v2 float64
	if se2, ok := obj2.(*Series); ok {
		key = -se2.ID
		v2 = se2.Get(0)
	} else if intVal, ok := obj2.(int); ok {
		key = intVal
		v2 = float64(intVal)
	} else if flt32Val, ok := obj2.(float32); ok {
		key = int(flt32Val * 100)
		v2 = float64(flt32Val)
	} else if fltVal, ok := obj2.(float64); ok {
		key = int(fltVal * 100)
		v2 = fltVal
	} else {
		fmt.Printf("invalid val for Series.objVal: %t", obj2)
		panic(ErrInvalidSeriesVal)
	}
	var newData = false
	var log *CrossLog
	if val, ok := se.XLogs[key]; ok {
		log = val
		if env.TimeStart > log.Time {
			newData = true
			log.Time = env.TimeStart
		}
	} else {
		newData = true
		log = &CrossLog{env.TimeStart, math.NaN(), []*XState{}}
		se.XLogs[key] = log
	}
	if newData {
		diffVal := se.Get(0) - v2
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
	}
	if len(log.Hist) > 0 {
		state := log.Hist[len(log.Hist)-1]
		return state.Sign * (env.BarNum - state.BarNum + 1)
	}
	return 0
}
