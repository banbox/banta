package ta

import (
	"github.com/banbox/banta"
)

// 本文件包含对 github.com/banbox/banta 包的封装函数。
// 这些封装函数旨在与 gopy 兼容，gopy 要求函数要么只有一个返回值，
// 要么有两个返回值，其中第二个必须是 error 类型。
// 因此，原始包中返回多个 *Series 的函数被修改为返回一个固定大小的 *Series 数组。

// --- 封装函数 ---
type Series = banta.Series

type BarEnv = banta.BarEnv

type CrossLog = banta.CrossLog

func AvgPrice(e *BarEnv) Series {
	return *banta.AvgPrice(e)
}

func HL2(h, l *Series) Series {
	return *banta.HL2(h, l)
}

func HLC3(h, l, c *Series) Series {
	return *banta.HLC3(h, l, c)
}

func Sum(obj *Series, period int) Series {
	return *banta.Sum(obj, period)
}

func SMA(obj *Series, period int) Series {
	return *banta.SMA(obj, period)
}

func VWMA(price *Series, vol *Series, period int) Series {
	return *banta.VWMA(price, vol, period)
}

func EMA(obj *Series, period int) Series {
	return *banta.EMA(obj, period)
}

func EMABy(obj *Series, period int, initType int) Series {
	return *banta.EMABy(obj, period, initType)
}

func RMA(obj *Series, period int) Series {
	return *banta.RMA(obj, period)
}

func RMABy(obj *Series, period int, initType int, initVal float64) Series {
	return *banta.RMABy(obj, period, initType, initVal)
}

func WMA(obj *Series, period int) Series {
	return *banta.WMA(obj, period)
}

func HMA(obj *Series, period int) Series {
	return *banta.HMA(obj, period)
}

func TR(high *Series, low *Series, close *Series) Series {
	return *banta.TR(high, low, close)
}

func ATR(high *Series, low *Series, close *Series, period int) Series {
	return *banta.ATR(high, low, close, period)
}

func MACD(obj *Series, fast int, slow int, smooth int) [2]Series {
	s1, s2 := banta.MACD(obj, fast, slow, smooth)
	return [2]Series{*s1, *s2}
}

func MACDBy(obj *Series, fast int, slow int, smooth int, initType int) [2]Series {
	s1, s2 := banta.MACDBy(obj, fast, slow, smooth, initType)
	return [2]Series{*s1, *s2}
}

func RSI(obj *Series, period int) Series {
	return *banta.RSI(obj, period)
}

func RSI50(obj *Series, period int) Series {
	return *banta.RSI50(obj, period)
}

func CRSI(obj *Series, period, upDn, roc int) Series {
	return *banta.CRSI(obj, period, upDn, roc)
}

func CRSIBy(obj *Series, period, upDn, roc, vtype int) Series {
	return *banta.CRSIBy(obj, period, upDn, roc, vtype)
}

func UpDown(obj *Series, vtype int) Series {
	return *banta.UpDown(obj, vtype)
}

func PercentRank(obj *Series, period int) Series {
	return *banta.PercentRank(obj, period)
}

func Highest(obj *Series, period int) Series {
	return *banta.Highest(obj, period)
}

func HighestBar(obj *Series, period int) Series {
	return *banta.HighestBar(obj, period)
}

func Lowest(obj *Series, period int) Series {
	return *banta.Lowest(obj, period)
}

func LowestBar(obj *Series, period int) Series {
	return *banta.LowestBar(obj, period)
}

func KDJ(high *Series, low *Series, close *Series, period int, sm1 int, sm2 int) [3]Series {
	k, d, j := banta.KDJ(high, low, close, period, sm1, sm2)
	return [3]Series{*k, *d, *j}
}

func KDJBy(high *Series, low *Series, close *Series, period int, sm1 int, sm2 int, maBy string) [3]Series {
	k, d, j := banta.KDJBy(high, low, close, period, sm1, sm2, maBy)
	return [3]Series{*k, *d, *j}
}

func Stoch(high, low, close *Series, period int) Series {
	return *banta.Stoch(high, low, close, period)
}

func MOM(obj *Series, period int) Series            { return *banta.MOM(obj, period) }
func OBV(close, volume *Series) Series              { return *banta.OBV(close, volume) }
func DEMA(obj *Series, period int) Series           { return *banta.DEMA(obj, period) }
func T3(obj *Series, period int) Series             { return *banta.T3(obj, period) }
func AroonOsc(high, low *Series, period int) Series { return *banta.AroonOsc(high, low, period) }
func AROONOSC(high, low *Series, period int) Series { return *banta.AROONOSC(high, low, period) }
func StochF(high, low, close *Series, period int, smooth ...int) [2]Series {
	k, d := banta.StochF(high, low, close, period, smooth...)
	return [2]Series{*k, *d}
}
func STOCHF(close, high, low *Series, period int, smooth ...int) [2]Series {
	k, d := banta.STOCHF(close, high, low, period, smooth...)
	return [2]Series{*k, *d}
}
func ULTOSC(high, low, close *Series, shortPeriod, mediumPeriod, longPeriod int) Series {
	return *banta.ULTOSC(high, low, close, shortPeriod, mediumPeriod, longPeriod)
}
func SAR(high, low *Series, step, max float64) Series  { return *banta.SAR(high, low, step, max) }
func PSAR(high, low *Series, step, max float64) Series { return *banta.PSAR(high, low, step, max) }

func Aroon(high *Series, low *Series, period int) [3]Series {
	s1, s2, s3 := banta.Aroon(high, low, period)
	return [3]Series{*s1, *s2, *s3}
}

func StdDev(obj *Series, period int) Series {
	return *banta.StdDev(obj, period)
}

func StdDevBy(obj *Series, period int, ddof int) [2]Series {
	s1, s2 := banta.StdDevBy(obj, period, ddof)
	return [2]Series{*s1, *s2}
}

func WrapFloatArr(res *Series, period int, inVal float64) []float64 {
	return banta.WrapFloatArr(res, period, inVal)
}

func BBANDS(obj *Series, period int, stdUp, stdDn float64) [3]Series {
	s1, s2, s3 := banta.BBANDS(obj, period, stdUp, stdDn)
	return [3]Series{*s1, *s2, *s3}
}

// Extra indicators keep the streaming API available to the Python binding.
func ROCR(obj *Series, period int) Series              { return *banta.ROCR(obj, period) }
func NATR(high, low, close *Series, period int) Series { return *banta.NATR(high, low, close, period) }
func TRIMA(obj *Series, period int) Series             { return *banta.TRIMA(obj, period) }
func SWMA(obj *Series) Series                          { return *banta.SWMA(obj) }
func ZLMA(obj *Series, period int) Series              { return *banta.ZLMA(obj, period) }
func PivotHigh(obj *Series, left, right int) Series    { return *banta.PivotHigh(obj, left, right) }
func PivotLow(obj *Series, left, right int) Series     { return *banta.PivotLow(obj, left, right) }

func TD(obj *Series) Series {
	return *banta.TD(obj)
}

func ADX(high *Series, low *Series, close *Series, period int) Series {
	return *banta.ADX(high, low, close, period)
}

func ADXBy(high *Series, low *Series, close *Series, period, smoothing, method int) Series {
	return *banta.ADXBy(high, low, close, period, smoothing, method)
}

func PluMinDI(high *Series, low *Series, close *Series, period int) [2]Series {
	s1, s2 := banta.PluMinDI(high, low, close, period)
	return [2]Series{*s1, *s2}
}

func PluMinDM(high *Series, low *Series, close *Series, period int) [2]Series {
	s1, s2 := banta.PluMinDM(high, low, close, period)
	return [2]Series{*s1, *s2}
}

func ROC(obj *Series, period int) Series {
	return *banta.ROC(obj, period)
}

func HeikinAshi(e *BarEnv) [4]Series {
	o, h, l, c := banta.HeikinAshi(e)
	return [4]Series{*o, *h, *l, *c}
}

func ER(obj *Series, period int) Series {
	return *banta.ER(obj, period)
}

func AvgDev(obj *Series, period int) Series {
	return *banta.AvgDev(obj, period)
}

func CCI(obj *Series, period int) Series {
	return *banta.CCI(obj, period)
}

func CMF(env *BarEnv, period int) Series {
	return *banta.CMF(env, period)
}

func ADL(env *BarEnv) Series {
	return *banta.ADL(env)
}

func ChaikinOsc(env *BarEnv, sml int, big int) Series {
	return *banta.ChaikinOsc(env, sml, big)
}

func AO(high, low *Series, fast, slow int) Series { return *banta.AO(high, low, fast, slow) }
func ADOSC(env *BarEnv, fast, slow int) Series    { return *banta.ADOSC(env, fast, slow) }
func EFI(env *BarEnv, period int) Series          { return *banta.EFI(env, period) }
func Donchian(high, low *Series, period int) [3]Series {
	u, m, d := banta.Donchian(high, low, period)
	return [3]Series{*u, *m, *d}
}
func Squeeze(high, low, close *Series, period int) Series {
	return *banta.Squeeze(high, low, close, period)
}
func Slope(obj *Series, period int) Series    { return *banta.Slope(obj, period) }
func TRIX(obj *Series, period int) Series     { return *banta.TRIX(obj, period) }
func TSI(obj *Series, short, long int) Series { return *banta.TSI(obj, short, long) }
func KST(obj *Series, r1, r2, r3, r4, s1, s2, s3, s4 int) Series {
	return *banta.KST(obj, r1, r2, r3, r4, s1, s2, s3, s4)
}
func DonchianPBand(high, low, close *Series, period int) Series {
	return *banta.DonchianPBand(high, low, close, period)
}
func KeltnerWBand(high, low, close *Series, period int, mult float64) Series {
	return *banta.KeltnerWBand(high, low, close, period, mult)
}
func VPCI(close, volume *Series, period int) Series { return *banta.VPCI(close, volume, period) }
func WilliamsPercent(high, low, close *Series, period int) Series {
	return *banta.WilliamsPercent(high, low, close, period)
}
func DX(high, low, close *Series, period int) Series { return *banta.DX(high, low, close, period) }
func Ichimoku(high, low, close *Series, conversion, base, span int) [5]Series {
	a, b, c, d, e := banta.Ichimoku(high, low, close, conversion, base, span)
	return [5]Series{*a, *b, *c, *d, *e}
}
func MAMA(obj *Series, fast, slow float64) [2]Series {
	a, b := banta.MAMA(obj, fast, slow)
	return [2]Series{*a, *b}
}
func Fisher(high, low *Series, period int) Series { return *banta.Fisher(high, low, period) }
func Correlation(a, b *Series, period int) Series { return *banta.Correlation(a, b, period) }

func KAMA(obj *Series, period int) Series {
	return *banta.KAMA(obj, period)
}

func KAMABy(obj *Series, period int, fast, slow int) Series {
	return *banta.KAMABy(obj, period, fast, slow)
}

func WillR(e *BarEnv, period int) Series {
	return *banta.WillR(e, period)
}

func StochRSI(obj *Series, rsiLen int, stochLen int, maK int, maD int) [2]Series {
	k, d := banta.StochRSI(obj, rsiLen, stochLen, maK, maD)
	return [2]Series{*k, *d}
}

func MFI(e *BarEnv, period int) Series {
	return *banta.MFI(e, period)
}

func RMI(obj *Series, period int, montLen int) Series {
	return *banta.RMI(obj, period, montLen)
}

func LinReg(obj *Series, period int) Series {
	return *banta.LinReg(obj, period)
}

func LinRegAdv(obj *Series, period int, angle, intercept, degrees, r, slope, tsf bool) Series {
	return *banta.LinRegAdv(obj, period, angle, intercept, degrees, r, slope, tsf)
}

func LinearRegAngle(obj *Series, period int) Series { return *banta.LinearRegAngle(obj, period) }

func LINEARREG_ANGLE(obj *Series, period int) Series { return *banta.LINEARREG_ANGLE(obj, period) }

func DPO(obj *Series, period int) Series { return *banta.DPO(obj, period) }

func CTI(obj *Series, period int) Series {
	return *banta.CTI(obj, period)
}

func CMO(obj *Series, period int) Series {
	return *banta.CMO(obj, period)
}

func CMOBy(obj *Series, period int, maType int) Series {
	return *banta.CMOBy(obj, period, maType)
}

func CHOP(e *BarEnv, period int) Series {
	return *banta.CHOP(e, period)
}

func ALMA(obj *Series, period int, sigma, distOff float64) Series {
	return *banta.ALMA(obj, period, sigma, distOff)
}

func Stiffness(obj *Series, maLen, stiffLen, stiffMa int) Series {
	return *banta.Stiffness(obj, maLen, stiffLen, stiffMa)
}

func DV2(h, l, c *Series, period, maLen int) Series {
	return *banta.DV2(h, l, c, period, maLen)
}

func UTBot(c, atr *Series, rate float64) Series {
	return *banta.UTBot(c, atr, rate)
}

func STC(obj *Series, period, fast, slow int, alpha float64) Series {
	return *banta.STC(obj, period, fast, slow, alpha)
}
