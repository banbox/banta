package banta

import "errors"

var (
	ErrInvalidSeriesVal = errors.New("invalid val for Series")
	ErrGetDataOfMerged  = errors.New("try get Data of merged series var")
	ErrRepeatAppend     = errors.New("repeat append on Series")
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
	Open       *Series
	High       *Series
	Low        *Series
	Close      *Series
	Volume     *Series
	Info       *Series
	Items      map[string]*Series
	XLogs      map[string]*CrossLog
	Data       map[string]interface{}
}

type Series struct {
	Env  *BarEnv
	Data []float64
	Cols []*Series
	Key  string
	Time int64
	More interface{}
}

type CrossLog struct {
	Key     string
	PrevVal float64
	Hist    []*XState // 正数表示上穿，负数下穿，绝对值表示BarNum
}

type XState struct {
	Sign   int
	BarNum int
}
