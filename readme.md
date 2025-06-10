# BanTA Technical Analysis Library
[中文文档](./readme.cn.md)  
**banta** is a high-performance technical analysis indicator library that supports both state-caching and parallel computation modes. It aims to provide a **highly flexible, high-performance, and user-friendly** indicator framework.

[![DeepWiki Q&A](https://deepwiki.com/badge.svg)](https://deepwiki.com/banbox/banta)

* **State-Caching Mode:** Updates and caches on each candle, eliminating the need to recalculate historical data. Indicator results are globally reused, similar to TradingView.
* **Parallel Computation Mode:** Computes all candles at once without caching. New candles require a full recalculation, similar to TA-Lib.
* **NaN Compatibility:** Intelligently skips NaN values in input data, resuming calculations with the previous state.
* **Rigorous Testing:** Each indicator is validated with unit tests under various conditions and compared against results from common indicator libraries.
* **Lightweight & Dependency-Free:** Pure Go implementation with zero external dependencies.
* **Python Support:** Packaged as the `bbta` module via gopy, ready to be imported and used directly in Python.

## Supported Indicators
#### Consistency Comparison with Common Indicator Platforms
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
--  This platform does not have the indicator  
✔  Consistent with this platform's results  
✔~ Mostly consistent with this platform's results (minor deviations)  
Ti Inconsistent with this platform's results  
```  

## How to Use (State-Caching Mode)
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

### Core Concept
Traditional technical analysis libraries like `TA-Lib` and `Pandas-TA` are widely used and highly optimized for performance, making them extremely fast when computing hundreds or thousands of candles at once.  
However, when your trading bot uses these libraries in live trading, every new candle requires passing in hundreds of historical candles. If you're running multiple symbols or operating on 1-minute or even 1-second timeframes, the computational delay can become unbearable.  
Many are familiar with TradingView, which uses Pine Script—an event-driven technical analysis engine. It doesn't recalculate historical candles upon receiving a new one but instead reuses cached results.  
This is the philosophy behind `BanTA`: event-driven computation, processing each candle as it arrives while leveraging cached results.

#### How State Caching Works in BanTA
In `BanTA`, state caching revolves around the `Series` sequence type. Most return values are sequences, and the `Series` struct includes a `Data []float64` field that records the indicator's values across candles.  
For example, `e.Close` is the closing price sequence, and `e.Close.Get(0)` retrieves the current closing price as a `float64`.  
Calculating a moving average is straightforward: `ma5 := ta.SMA(e.Close, 5)`, which returns another sequence.  
Some indicators like KDJ return multiple fields: `kdjRes := ta.KDJ(e.High, e.Low, e.Close, 9, 3, 3).Cols`, where `Cols` contains an array of sequences (e.g., K and D lines).

## How to Use (Parallel Computation)
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

### Note
For research purposes, we recommend using parallel computation to compute indicators in bulk. For live trading or event-driven backtesting, state-cached indicators offer higher efficiency.

## Python Installation  
```shell  
pip install bbta  
```  
Only supports Python 8 and above. Currently not compatible with Python 13 on macOS and Windows.

## Python Usage (State-Caching Mode)
```python
from bbta import ta

# 1. Create an environment
# BarEnv manages state; create one for each time frame/trading pair.
env = ta.BarEnv(TimeFrame="1m")

# 2. Prepare candle data
# (timestamp ms, open, high, low, close, volume)
klines = [
    (1672531200000, 100, 102, 99, 101, 1000), (1672531260000, 101, 103, 100, 102, 1200),
    (1672531320000, 102, 105, 101, 104, 1500), (1672531380000, 104, 105, 103, 103, 1300),
    (1672531440000, 103, 104, 102, 103, 1100), (1672531500000, 103, 106, 103, 105, 1600),
    (1672531560000, 105, 107, 104, 106, 1800), (1672531620000, 106, 106, 102, 103, 2000),
    (1672531680000, 103, 104, 101, 102, 1700), (1672531740000, 102, 103, 100, 101, 1400),
]

# 3. Simulate candle pushes
# In live trading, call OnBar for each new candle.
for kline in klines:
    ts, o, h, l, c, v = kline
    env.OnBar(ts, o, h, l, c, v, 0)

    # 4. Calculate indicators
    ma5 = ta.Series(ta.SMA(env.Close, 5))
    ma30 = ta.Series(ta.SMA(env.Close, 30))

    # Get the latest value
    ma5_val = ma5.Get(0)
    ma30_val = ma30.Get(0)
    print(f"Close={c:.2f}, MA5={ma5_val:.2f}, MA30={ma30_val:.2f}")

## Python Usage (Parallel Computation)
```python
from bbta import tav, go

# 1. Prepare data
# Functions in parallel mode accept go.Slice_float64 type.
# We can create it from a Python list.
high_py = [102.0, 103.0, 105.0, 105.0, 104.0, 106.0, 107.0, 106.0, 104.0, 103.0]
low_py = [99.0, 100.0, 101.0, 103.0, 102.0, 103.0, 104.0, 102.0, 101.0, 100.0]
close_py = [101.0, 102.0, 104.0, 103.0, 103.0, 105.0, 106.0, 103.0, 102.0, 101.0]

high = go.Slice_float64(high_py)
low = go.Slice_float64(low_py)
close = go.Slice_float64(close_py)

# 2. Compute all indicators at once
# The result is also a go.Slice type.
ma5 = tav.SMA(close, 5)
atr = tav.ATR(high, low, close, 14)

# 3. View the results
# You can convert it to a Python list to view.
print(f"Close: {list(close)[-5:]}")
print(f"MA5:   {[f'{x:.2f}' for x in list(ma5)[-5:]]}")
print(f"ATR:   {[f'{x:.2f}' for x in list(atr)[-5:]]}")

# For indicators with multiple return values, like KDJ
kdj_result = tav.KDJ(high, low, close, 9, 3, 3)
k_line = kdj_result[0]
d_line = kdj_result[1]
j_line = kdj_result[2]
print(f"K-line: {[f'{x:.2f}' for x in list(k_line)[-5:]]}")
