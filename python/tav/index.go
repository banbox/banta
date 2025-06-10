// Package tav provides a gopy-compatible wrapper for the github.com/banbox/banta/tav library.
// It modifies functions with multiple return values to return a single array of slices,
// making them suitable for Python bindings generation.
package tav

import (
	banta_tav "github.com/banbox/banta/tav"
)

// HL2 calculates the average of two series (e.g., high and low).
func HL2(a, b []float64) []float64 {
	return banta_tav.HL2(a, b)
}

// HLC3 calculates the average of three series (e.g., high, low, and close).
func HLC3(a, b, c []float64) []float64 {
	return banta_tav.HLC3(a, b, c)
}

// Sum calculates the rolling sum of a series over a given period.
func Sum(data []float64, period int) []float64 {
	return banta_tav.Sum(data, period)
}

// SMA calculates the Simple Moving Average.
func SMA(data []float64, period int) []float64 {
	return banta_tav.SMA(data, period)
}

// VWMA calculates the Volume Weighted Moving Average.
func VWMA(price []float64, volume []float64, period int) []float64 {
	return banta_tav.VWMA(price, volume, period)
}

// EMA calculates the Exponential Moving Average.
func EMA(data []float64, period int) []float64 {
	return banta_tav.EMA(data, period)
}

// EMABy calculates the Exponential Moving Average with a specified initialization type.
func EMABy(data []float64, period int, initType int) []float64 {
	return banta_tav.EMABy(data, period, initType)
}

// RMA calculates the Wilder's Smoothing Average (Relative Moving Average).
func RMA(data []float64, period int) []float64 {
	return banta_tav.RMA(data, period)
}

// RMABy calculates the Wilder's Smoothing Average with specified initialization.
func RMABy(data []float64, period int, initType int, initVal float64) []float64 {
	return banta_tav.RMABy(data, period, initType, initVal)
}

// WMA calculates the Weighted Moving Average.
func WMA(data []float64, period int) []float64 {
	return banta_tav.WMA(data, period)
}

// HMA calculates the Hull Moving Average.
func HMA(data []float64, period int) []float64 {
	return banta_tav.HMA(data, period)
}

// TR calculates the True Range.
func TR(high, low, close []float64) []float64 {
	return banta_tav.TR(high, low, close)
}

// ATR calculates the Average True Range.
func ATR(high, low, close []float64, period int) []float64 {
	return banta_tav.ATR(high, low, close, period)
}

// RSI calculates the Relative Strength Index.
func RSI(data []float64, period int) []float64 {
	return banta_tav.RSI(data, period)
}

// RSIBy calculates the Relative Strength Index with a subtraction value.
func RSIBy(data []float64, period int, subVal float64) []float64 {
	return banta_tav.RSIBy(data, period, subVal)
}

// MACD calculates the Moving Average Convergence Divergence.
// Returns [2][]float64{macd, signal}.
func MACD(data []float64, fast, slow, smooth int) [2][]float64 {
	macd, signal := banta_tav.MACD(data, fast, slow, smooth)
	return [2][]float64{macd, signal}
}

// MACDBy calculates the MACD with a specified initialization type.
// Returns [2][]float64{macd, signal}.
func MACDBy(data []float64, fast, slow, smooth, initType int) [2][]float64 {
	macd, signal := banta_tav.MACDBy(data, fast, slow, smooth, initType)
	return [2][]float64{macd, signal}
}

// StdDev calculates the Standard Deviation.
func StdDev(data []float64, period int) []float64 {
	return banta_tav.StdDev(data, period)
}

// StdDevBy calculates the Standard Deviation with a specified delta degrees of freedom (ddof).
// Returns [2][]float64{stddev, mean}.
func StdDevBy(data []float64, period int, ddof int) [2][]float64 {
	stddev, mean := banta_tav.StdDevBy(data, period, ddof)
	return [2][]float64{stddev, mean}
}

// BBANDS calculates Bollinger Bands.
// Returns [3][]float64{upperBand, middleBand, lowerBand}.
func BBANDS(data []float64, period int, stdUp, stdDn float64) [3][]float64 {
	upper, middle, lower := banta_tav.BBANDS(data, period, stdUp, stdDn)
	return [3][]float64{upper, middle, lower}
}

// TD calculates the Tom DeMark Sequential.
func TD(data []float64) []float64 {
	return banta_tav.TD(data)
}

// ADX calculates the Average Directional Index.
func ADX(high, low, close []float64, period int) []float64 {
	return banta_tav.ADX(high, low, close, period)
}

// ADXBy calculates the ADX with specified smoothing and method.
func ADXBy(high, low, close []float64, period, smoothing, method int) []float64 {
	return banta_tav.ADXBy(high, low, close, period, smoothing, method)
}

// PluMinDI calculates the Plus Directional Indicator (+DI) and Minus Directional Indicator (-DI).
// Returns [2][]float64{plusDI, minusDI}.
func PluMinDI(high, low, close []float64, period int) [2][]float64 {
	plusDI, minusDI := banta_tav.PluMinDI(high, low, close, period)
	return [2][]float64{plusDI, minusDI}
}

// PluMinDM calculates the Plus Directional Movement (+DM) and Minus Directional Movement (-DM).
// Returns [2][]float64{plusDM, minusDM}.
func PluMinDM(high, low, cls []float64, period int) [2][]float64 {
	plusDM, minusDM := banta_tav.PluMinDM(high, low, cls, period)
	return [2][]float64{plusDM, minusDM}
}

// Highest finds the highest value over a specified period.
func Highest(data []float64, period int) []float64 {
	return banta_tav.Highest(data, period)
}

// Lowest finds the lowest value over a specified period.
func Lowest(data []float64, period int) []float64 {
	return banta_tav.Lowest(data, period)
}

// HighestBar finds the offset to the highest value bar over a specified period.
func HighestBar(data []float64, period int) []float64 {
	return banta_tav.HighestBar(data, period)
}

// LowestBar finds the offset to the lowest value bar over a specified period.
func LowestBar(data []float64, period int) []float64 {
	return banta_tav.LowestBar(data, period)
}

// Stoch calculates the Stochastic Oscillator.
func Stoch(high, low, close []float64, period int) []float64 {
	return banta_tav.Stoch(high, low, close, period)
}

// KDJ calculates the KDJ indicator.
// Returns [3][]float64{k, d, j}.
func KDJ(high, low, close []float64, period, sm1, sm2 int) [3][]float64 {
	k, d, j := banta_tav.KDJ(high, low, close, period, sm1, sm2)
	return [3][]float64{k, d, j}
}

// KDJBy calculates the KDJ indicator with a specified MA type.
// Returns [3][]float64{k, d, rsv}.
func KDJBy(high, low, close []float64, period int, sm1 int, sm2 int, maBy string) [3][]float64 {
	k, d, rsv := banta_tav.KDJBy(high, low, close, period, sm1, sm2, maBy)
	return [3][]float64{k, d, rsv}
}

// Aroon calculates the Aroon Indicator.
// Returns [3][]float64{aroonUp, aroonDown, aroonOscillator}.
func Aroon(high, low []float64, period int) [3][]float64 {
	up, down, osc := banta_tav.Aroon(high, low, period)
	return [3][]float64{up, down, osc}
}

// ROC calculates the Rate of Change.
func ROC(data []float64, period int) []float64 {
	return banta_tav.ROC(data, period)
}

// UpDown classifies price movement as up, down, or flat.
func UpDown(data []float64, vtype int) []float64 {
	return banta_tav.UpDown(data, vtype)
}

// PercentRank calculates the percentile rank of the current value over a period.
func PercentRank(data []float64, period int) []float64 {
	return banta_tav.PercentRank(data, period)
}

// CRSI calculates the Connors RSI.
func CRSI(data []float64, period, upDn, rocVal int) []float64 {
	return banta_tav.CRSI(data, period, upDn, rocVal)
}

// CRSIBy calculates the Connors RSI with a specified type.
func CRSIBy(data []float64, period, upDn, rocVal, vtype int) []float64 {
	return banta_tav.CRSIBy(data, period, upDn, rocVal, vtype)
}

// ER calculates the Efficiency Ratio (Kaufman).
func ER(data []float64, period int) []float64 {
	return banta_tav.ER(data, period)
}

// AvgDev calculates the Average Deviation.
func AvgDev(data []float64, period int) []float64 {
	return banta_tav.AvgDev(data, period)
}

// CCI calculates the Commodity Channel Index.
func CCI(data []float64, period int) []float64 {
	return banta_tav.CCI(data, period)
}

// CMF calculates the Chaikin Money Flow.
func CMF(high, low, close, volume []float64, period int) []float64 {
	return banta_tav.CMF(high, low, close, volume, period)
}

// KAMA calculates the Kaufman's Adaptive Moving Average.
func KAMA(data []float64, period int) []float64 {
	return banta_tav.KAMA(data, period)
}

// KAMABy calculates the KAMA with fast and slow periods.
func KAMABy(data []float64, period int, fast, slow int) []float64 {
	return banta_tav.KAMABy(data, period, fast, slow)
}

// WillR calculates the Williams %R.
func WillR(high, low, close []float64, period int) []float64 {
	return banta_tav.WillR(high, low, close, period)
}

// StochRSI calculates the Stochastic RSI.
// Returns [2][]float64{k, d}.
func StochRSI(obj []float64, rsiLen int, stochLen int, maK int, maD int) [2][]float64 {
	k, d := banta_tav.StochRSI(obj, rsiLen, stochLen, maK, maD)
	return [2][]float64{k, d}
}

// MFI calculates the Money Flow Index.
func MFI(high, low, close, volume []float64, period int) []float64 {
	return banta_tav.MFI(high, low, close, volume, period)
}

// RMI calculates the Relative Momentum Index.
func RMI(data []float64, period int, montLen int) []float64 {
	return banta_tav.RMI(data, period, montLen)
}

// LinReg calculates the Linear Regression.
func LinReg(data []float64, period int) []float64 {
	return banta_tav.LinReg(data, period)
}

// LinRegAdv calculates advanced Linear Regression values.
func LinRegAdv(data []float64, period int, angle, intercept, degrees, r, slope, tsf bool) []float64 {
	return banta_tav.LinRegAdv(data, period, angle, intercept, degrees, r, slope, tsf)
}

// CTI calculates the Correlation Trend Indicator.
func CTI(data []float64, period int) []float64 {
	return banta_tav.CTI(data, period)
}

// CMO calculates the Chande Momentum Oscillator.
func CMO(data []float64, period int) []float64 {
	return banta_tav.CMO(data, period)
}

// CMOBy calculates the CMO with a specified MA type.
func CMOBy(data []float64, period int, maType int) []float64 {
	return banta_tav.CMOBy(data, period, maType)
}

// CHOP calculates the Choppiness Index.
func CHOP(high, low, close []float64, period int) []float64 {
	return banta_tav.CHOP(high, low, close, period)
}

// ALMA calculates the Arnaud Legoux Moving Average.
func ALMA(data []float64, period int, sigma, distOff float64) []float64 {
	return banta_tav.ALMA(data, period, sigma, distOff)
}

// Stiffness calculates the Stiffness indicator.
func Stiffness(data []float64, maLen, stiffLen, stiffMa int) []float64 {
	return banta_tav.Stiffness(data, maLen, stiffLen, stiffMa)
}

// DV2 calculates the DV2 indicator.
func DV2(h, l, c []float64, period, maLen int) []float64 {
	return banta_tav.DV2(h, l, c, period, maLen)
}

// UTBot calculates the UT Bot indicator.
func UTBot(c, atr []float64, rate float64) []float64 {
	return banta_tav.UTBot(c, atr, rate)
}

// STC calculates the Schaff Trend Cycle.
func STC(data []float64, period, fast, slow int, alpha float64) []float64 {
	return banta_tav.STC(data, period, fast, slow, alpha)
}

// HeikinAshi calculates Heikin-Ashi candlesticks.
// Returns [4][]float64{haOpen, haHigh, haLow, haClose}.
func HeikinAshi(open, high, low, close []float64) [4][]float64 {
	haOpen, haHigh, haLow, haClose := banta_tav.HeikinAshi(open, high, low, close)
	return [4][]float64{haOpen, haHigh, haLow, haClose}
}

// Cross detects crossovers between two data series.
func Cross(data1 []float64, data2 []float64) []int {
	return banta_tav.Cross(data1, data2)
}
