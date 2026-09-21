package banta

import "github.com/banbox/banta/tav"
import "math"
import "slices"

/*
CTI Correlation Trend Indicator

The Correlation Trend Indicator is an oscillator created by John Ehler in 2020.
It assigns a value depending on how close prices in that range are to following
a positively- or negatively-sloping straight line. Values range from -1 to 1.
This is a wrapper for LinRegAdv.

suggest period: 20
*/
func CTI(obj *Series, period int) *Series {
	return LinRegAdv(obj, period, false, false, false, true, false, false)
}

/*
LinReg Linear Regression Moving Average

Linear Regression Moving Average (LINREG). This is a simplified version of a
Standard Linear Regression. LINREG is a rolling regression of one variable. A
Standard Linear Regression is between two or more variables.
*/
func LinReg(obj *Series, period int) *Series {
	return LinRegAdv(obj, period, false, false, false, false, false, false)
}

// LINEARREG_ANGLE is the conventional TA-Lib spelling.
func LINEARREG_ANGLE(obj *Series, period int) *Series {
	return LinearRegAngle(obj, period)
}

// LinearRegAngle returns the TA-Lib LINEARREG_ANGLE value in degrees.
// It is atan of the least-squares slope over the most recent period bars;
// the first period-1 bars are NaN. This is equivalent to LinRegAdv with
// angle=true and degrees=true.
func LinearRegAngle(obj *Series, period int) *Series {
	return LinRegAdv(obj, period, true, false, true, false, false, false)
}

// Slope is the least-squares linear-regression slope over period bars.
func Slope(obj *Series, period int) *Series {
	return LinRegAdv(obj, period, false, false, false, false, true, false)
}

func Correlation(a, b *Series, period int) *Series {
	r := a.To("_correlation", period)
	if !r.Cached() {
		r.Append(last(tav.Correlation(hist(a), hist(b), period)))
	}
	return r
}

func Highest(obj *Series, period int) *Series {
	res := obj.To("_hh", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		inVal := obj.Get(0)
		if math.IsNaN(inVal) {
			res.Append(math.NaN())
		} else {
			// 获取周期内的数据
			values := WrapFloatArr(res, period, inVal)
			if len(values) < period {
				res.Append(math.NaN())
			} else {
				res.Append(slices.Max(values))
			}
		}
	}
	return res
}

func HighestBar(obj *Series, period int) *Series {
	res := obj.To("_hhb", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		if math.IsNaN(obj.Get(0)) {
			res.Append(math.NaN())
		} else {
			values, ids := obj.RangeValid(0, period)
			if len(values) < period {
				res.Append(math.NaN())
			} else {
				maxIdx, maxVal := -1, math.NaN()

				// 遍历以寻找非 NaN 的最大值
				for i, v := range values {
					// 如果 maxVal 是 NaN (说明这是第一个有效值) 或当前值更大，则更新
					if maxIdx < 0 || v > maxVal {
						maxVal = v
						maxIdx = ids[i]
					}
				}
				res.Append(maxIdx)
			}
		}
	}
	return res
}

func LinRegAdv(obj *Series, period int, angle, intercept, degrees, r, slope, tsf bool) *Series {
	hash := period*100 + boolToHash(angle, intercept, degrees, r, slope, tsf)
	res := obj.To("_linreg", hash)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		sumY := Sum(obj, period).Get(0)
		val := obj.Get(0)
		if math.IsNaN(val) {
			res.Append(math.NaN())
		} else {
			arr := WrapFloatArr(res, period, val)
			if len(arr) < period || math.IsNaN(sumY) {
				res.Append(math.NaN())
			} else {
				periodF := float64(period)
				var sumXY = float64(0)
				var sumX = periodF * float64(period+1) * 0.5
				sumY2 := float64(0)
				for i := 0; i < period; i++ {
					v := arr[i]
					sumXY += float64(i+1) * v
					if r {
						sumY2 += v * v
					}
				}
				sumX2 := sumX * (2*periodF + 1) / 3
				divisor := periodF*sumX2 - sumX*sumX
				m := (periodF*sumXY - sumX*sumY) / divisor
				if slope {
					res.Append(m)
					return res
				}
				b := (sumY*sumX2 - sumX*sumXY) / divisor
				if intercept {
					res.Append(b)
					return res
				}
				if angle {
					theta := math.Atan(m)
					if degrees {
						theta *= 180 / math.Pi
					}
					res.Append(theta)
					return res
				}
				if r {
					rn := periodF*sumXY - sumX*sumY
					rd := math.Pow(divisor*(periodF*sumY2-sumY*sumY), 0.5)
					res.Append(rn / rd)
					return res
				}
				if tsf {
					res.Append(m*periodF + b)
				} else {
					res.Append(m*(periodF-1) + b)
				}
			}
		}
	}
	return res
}

func Lowest(obj *Series, period int) *Series {
	res := obj.To("_ll", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		inVal := obj.Get(0)
		if math.IsNaN(inVal) {
			res.Append(math.NaN())
		} else {
			// 获取周期内的数据
			values := WrapFloatArr(res, period, inVal)
			if len(values) < period {
				res.Append(math.NaN())
			} else {
				res.Append(slices.Min(values))
			}
		}
	}
	return res
}

func LowestBar(obj *Series, period int) *Series {
	res := obj.To("_llb", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		if math.IsNaN(obj.Get(0)) {
			res.Append(math.NaN())
		} else {
			values, ids := obj.RangeValid(0, period)
			if len(values) < period {
				res.Append(math.NaN())
			} else {
				minIdx, minVal := -1, math.NaN()

				// 遍历以寻找非 NaN 的最大值
				for i, v := range values {
					// 如果 maxVal 是 NaN (说明这是第一个有效值) 或当前值更大，则更新
					if minIdx < 0 || v < minVal {
						minVal = v
						minIdx = ids[i]
					}
				}
				res.Append(minIdx)
			}
		}
	}
	return res
}
