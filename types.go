package banta

import (
	"errors"
	"sync"
)

var (
	ErrInvalidSeriesVal = errors.New("invalid val for Series")
)

type Kline struct {
	Time      int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Quote     float64 // volume in quote
	BuyVolume float64 // taker buy volume
	TradeNum  int64
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
	Quote      *Series
	BuyVolume  *Series
	TradeNum   *Series
	Data       sync.Map // map[string]interface{}
	Items      map[int]*Series
	Lock       sync.Mutex // 用户手动控制的锁，用于保护并发访问
}

type Series struct {
	ID      int
	Env     *BarEnv
	Data    []float64
	Cols    []*Series
	Time    int64
	More    interface{}
	DupMore func(interface{}) interface{}
	Subs    map[string]map[int]*Series // 由此序列派生的；function：hash：object
	XLogs   map[int]*CrossLog          // 此序列交叉记录
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
