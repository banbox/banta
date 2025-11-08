package banta

import (
	"errors"
	"sync"
)

var (
	ErrInvalidSeriesVal = errors.New("invalid val for Series")
)

type Kline struct {
	Time   int64
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
	Info   float64
}

type BarEnv struct {
	TimeStart  int64
	TimeStop   int64
	Exchange   string
	MarketType string
	Symbol     string
	TimeFrame  string
	TFMSecs    int64 //周期的毫秒间隔
	BarNum     int
	MaxCache   int
	VNum       int
	Open       *Series
	High       *Series
	Low        *Series
	Close      *Series
	Volume     *Series
	Info       *Series
	Data       map[string]interface{}
	Items      map[int]*Series
	LockData   sync.RWMutex
	LockItems  sync.RWMutex
}

type Series struct {
	ID         int
	Env        *BarEnv
	Data       []float64
	Cols       []*Series
	Time       int64
	More       interface{}
	DupMore    func(interface{}) interface{}
	Subs       map[string]map[int]*Series // 由此序列派生的；function：hash：object
	XLogs      map[int]*CrossLog          // 此序列交叉记录
	LockSubMap map[string]*sync.Mutex
	LockSub    sync.Mutex
	LockXLogs  sync.Mutex
	LockData   sync.RWMutex
}

type CrossLog struct {
	Time    int64
	PrevVal float64
	Hist    []*XState // 正数表示上穿，负数下穿，绝对值表示BarNum
}

type XState struct {
	Sign   int
	BarNum int
}
