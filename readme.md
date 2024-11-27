# Ban Technical Analysis
[中文文档](./readme.cn.md)  

**banta** is an event-based technical analysis framework. It updates and caches indicators on each K-line, and the indicator calculation results are reused globally. It aims to provide a **high degree of freedom**, **high performance**, **simple and easy-to-use** indicator framework.

## Core Concept
Traditional technical analysis libraries such as `ta-lib` and `pandas-ta` are very popular and have undergone a lot of performance optimization. They are very fast when calculating hundreds or thousands of candlesticks at a time.

However, when using these libraries in your robot's real-time trading, you need to pass the data of the previous hundreds of candlesticks at the same time when you receive a new candlestick. If you are running more symbols or if you are running on 1m or even 1s, the calculation delay will be too large to be tolerated.

Many people have used TradingView. The PineScript it uses is an event-based technical analysis engine. It does not re-run the previous candlesticks when receiving a new candlestick, but uses the cached results from the previous candlesticks.

This is also the design concept of `BanTA`, based on events, running once for each candlestick, and using the cached results from the previous candlesticks.

### How BanTA Works
In BanTA, `Series` sequence is the key type. Basically, all return values are sequences. The `Series` has the `Data []float64` field, which records the values of the current indicator for each candlestick.

The most commonly used `e.Close` is the closing price sequence. `e.Close.Get(0)` is the current closing price of type `float64`;

Calculating the average is also simple: `ma5 := ta.SMA(e.Close, 5)`, the returned ma5 is also a sequence;

Some indicators such as KDJ generally return two fields k and d `kdjRes := ta.KDJ(e.High, e.Low, e.Close, 9, 3, 3).Cols`, you can get an array of two sequences from Cols.

## Supported indicators
#### Comparison of consistency with common indicator platform results
| banta       | MyTT | TA-lib Class | TA-lib Metastock | Pandas-TA | TradingView |
|-------------|:----:|:------------:|:----------------:|:---------:|:-----------:| 
| AvgPrice    |  ✔   |      ✔       |        ✔         |     ✔     |      ✔      |
| Sum         |  ✔   |      ✔       |        ✔         |     ✔     |      ✔      |
| SMA         |  T1  |      ✔       |        ✔         |     ✔     |      ✔      |
| EMA         |  T1  |      T1      |        ✔         |     ✔     |     T2      |
| EMABy1      |  ✔   |      T1      |        T2        |    T2     |     T3      |
| RMA         |  --  |      --      |        --        |    T1     |     --      |
| VWMA        |  --  |      --      |        --        |     ✔     |      ✔      |
| WMA         |  ✔   |      ✔       |        ✔         |     ✔     |      ✔      |
| HMA         |  --  |      --      |        --        |     ✔     |      ✔      |
| TR          |  --  |      ✔       |        ✔         |     ✔     |     --      |
| ATR         |  T1  |      ✔       |        ✔         |    T2     |     T3      |
| MACD        |  T1  |      T2      |        T1        |     ✔     |     T3      |
| RSI         |  T1  |      ✔       |        ✔         |    T2     |     T3      |
| KDJ         |  T1  |      T2      |        T1        |    T3     |      ✔      |
| Stoch       |  --  |      ✔       |        ✔         |    --     |      ✔      |
| BBANDS      |  ✔   |      ✔       |        ✔         |     ✔     |      ✔      |
| Aroon       |  --  |      ✔       |        ✔         |     ✔     |     T1      |
| ADX         |  --  |      ✔       |        ✔         |    T1     |     T2      |
| ADXBy1      |  --  |      T1      |        T1        |    T2     |      ✔      |
| PluMinDI    |  --  |      ✔       |        ✔         |    --     |     --      |
| PluMinDM    |  --  |      ✔       |        ✔         |    --     |     --      |
| ROC         |  ✔   |      ✔       |        ✔         |    --     |      ✔      |
| TNR/ER      |  --  |      --      |        --        |    --     |     --      |
| CCI         |  ✔   |      ✔       |        ✔         |     ✔     |      ✔      |
| CMF         |  --  |      --      |        --        |     ✔     |      ✔      |
| KAMA        |  --  |      ✔~      |        ✔~        |    T1     |     ✔~      |
| WillR       |  --  |      ✔       |        ✔         |     ✔     |      ✔      |
| StochRSI    |  --  |      ✔       |        ✔         |     ✔     |     ✔~      |
| MFI         |  ✔   |      ✔       |        ✔         |     ✔     |      ✔      |
| RMI         |  --  |      --      |        --        |    --     |     ✔~      |
| CTI         |  --  |      --      |        --        |     ✔     |     T1      |
| LinReg      |  --  |      --      |        --        |     ✔     |      ?      |
| CMO         |  --  |      ✔       |        ✔         |     ✔     |     T1      |
| CMOBy1      |  --  |      T1      |        T1        |    T1     |      ✔      |
| CHOP        |  --  |      --      |        --        |     ✔     |     T1      |
| ALMA        |  --  |      --      |        --        |     ✔     |     T1      |
| Stiffness   |  --  |      --      |        --        |    --     |      ✔      |
| PercentRank |  --  |      --      |        --        |    --     |     ✔~      |
| CRSI        |  --  |      --      |        --        |    --     |     ✔~      |
| CRSIBy1     |  --  |  community   |        --        |    --     |     --      |
```text
-- This platform does not have this indicator
✔ The calculation results are consistent with this platform
✔~ The calculation results are basically consistent with this platform (with some deviation)
Ti The calculation results are inconsistent with this platform
```

## How to Use
```go
package main
import (
	"fmt"
	ta "github.com/banbox/banta"
)

var envMap = make(map[string]*ta.BarEnv)

func OnBar(symbol string, timeframe string, bar *ta.Kline) {
	envKey := fmt.Sprintf("%s_%s", symbol, timeframe)
	e, ok := envMap[envKey]
	if !ok {
		e = &ta.BarEnv{
			TimeFrame: timeframe,
			BarNum:    1,
		}
		envMap[envKey] = e
	}
	e.OnBar(bar)
	ma5 := ta.SMA(e.Close, 5)
	ma30 := ta.SMA(e.Close, 30)
	atr := ta.ATR(e.High, e.Low, e.Close, 14).Get(0)
	xnum := ta.Cross(ma5, ma30)
	if xnum == 1 {
		// ma5 cross up ma30
		curPrice := e.Close.Get(0) // or bar.Close
		stopLoss := curPrice - atr
		fmt.Printf("open long at %f, stoploss: %f", curPrice, stopLoss)
	} else if xnum == -1 {
		// ma5 cross down ma30
		curPrice := e.Close.Get(0)
		fmt.Printf("close long at %f", curPrice)
	}
	kdjRes := ta.KDJ(e.High, e.Low, e.Close, 9, 3, 3).Cols
	k, d := kdjRes[0], kdjRes[1]
}
```