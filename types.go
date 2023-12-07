package banta

import "errors"

var (
	ErrInvalidSeriesVal = errors.New("invalid val for Series")
	ErrGetDataOfMerged  = errors.New("try get Data of merged series var")
	ErrRepeatAppend     = errors.New("repeat append on Series")
)

type Kline struct {
	Time   int
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

type BarEnv struct {
	TimeStart  int
	TimeStop   int
	Exchange   string
	MarketType string
	TimeFrame  string
	TFMSecs    int //周期的毫秒间隔
	BarNum     int
	MaxCache   int
	Open       *Series
	High       *Series
	Low        *Series
	Close      *Series
	Volume     *Series
	Items      map[string]*Series
	XLogs      map[string]*CrossLog
}

type Series struct {
	Env  *BarEnv
	Data []float64
	Cols []*Series
	Key  string
	Time int
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
