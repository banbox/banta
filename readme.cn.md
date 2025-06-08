# BanTA 技术分析库
**banta**是一个高性能的技术分析指标库，支持状态缓存/并行计算两种模式。旨在提供**高自由度、高性能、简洁易用**的指标框架。

[![DeepWiki问答](https://deepwiki.com/badge.svg)](https://deepwiki.com/banbox/banta)

* **状态缓存模式：** 在每个K线上更新并缓存，无需重新计算历史数据，指标计算结果全局重用；类似TradingView
* **并行计算模式：** 一次性计算所有K线，无缓存，收到新K线需再次重新计算；类似ta-lib
* **nan值兼容：** 智能忽略输入数据中间的nan值，后续计算继续使用之前的状态
* **严格测试：** 每个指标都使用各种条件单元测试验证，并和常见指标库结果对比
* **轻量无依赖：** 仅使用golang，无任何依赖包

## 支持的指标
#### 和常见指标平台结果一致性对比
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
| KAMA        |  --  |      ✔       |        ✔         |    T1     |     ✔~      |
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
| DV          |  --  |      --      |        --        |    --     |     --      |
| UTBot       |  --  |      --      |        --        |    --     |      ✔      |
| STC         |  --  |      --      |        --        |    --     |      ✔      |
```text
-- 此平台没有该指标
✔  和此平台计算结果一致
✔~ 和此平台计算结果基本一致(有一定偏差)
Ti  和此平台计算结果不一致 
```

## 如何使用(带状态缓存)
```go
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
### 核心理念
传统的技术分析库比如`ta-lib`和`pandas-ta`这些库非常流行，也进行了很多性能优化，在一次性计算几百上千跟K线时速度非常快。  
但在你的机器人实盘运行使用这些库时，每次收到新的K线都要同时传入前面几百个K线数据，如果运行的标的更多一些，或者如果在1m甚至1s上运行时，计算延迟就会大到无法忍受。  
很多人都用过TradingView，它使用的PineScript就是一个基于事件的技术分析引擎；它不会在收到一根新K线时重新运行前面的K线，而是使用前面缓存的结果。  
这也是`BanTA`的设计理念，基于事件，每收到一个K线运行一次，使用前面缓存的结果。

#### BanTA中状态缓存是如何实现的
在`BanTA`中，状态缓存的核心是`Series`序列类型，基本上所有返回值都是序列，`Series`中有`Data []float64`字段，它记录了当前指标在每个K线的值。  
最常用的`e.Close`就是收盘价序列，`e.Close.Get(0)`就是`float64`类型的当前收盘价；  
计算均值也很简单：`ma5 := ta.SMA(e.Close, 5)`，返回的ma5也是一个序列；  
有些指标如KDJ一般返回k和d两个字段`kdjRes := ta.KDJ(e.High, e.Low, e.Close, 9, 3, 3).Cols`，可以从`Cols`中得到两个序列组成的数组。


## 如何使用（并行计算）
```go
import (
	"github.com/banbox/banta/tav"
)

func main(){
	highArr := []float64{1.01, 1.01, 1.02, 0.996, 0.98, 0.993, 0.99, 1.0, 1.02}
    lowArr := []float64{0.99, 1.0, 1.0, 0.98, 0.965, 0.98, 0.98, 0.984, 1.0}
	closeArr := []float64{1.0, 1.01, 1.0, 0.99, 0.97, 0.981, 0.988, 0.992, 1.002}
	sma := tav.SMA(closeArr, 5)
    ma30 := tav.SMA(closeArr, 30)
    atr := tav.ATR(highArr, lowArr, closeArr, 14)
    xArr := tav.Cross(ma5, ma30)
	
}
```
### 提示
建议您在进行研究时使用并行计算一次性获取指标计算结果，在实盘或基于事件驱动回测时使用带状态指标，效率更高。
