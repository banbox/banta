package tav

import "math"

// CTI Correlation Trend Indicator 并行版本
// CTI是一个由John Ehler在2020年创建的振荡器
// 它根据价格范围内的价格接近正斜率或负斜率直线的程度来分配值
// 值范围从-1到1，建议周期：20
// 当前是性能优化版本，快20%；等同于LinRegAdv(data, period, false, false, false, true, false, false)
func CTI(data []float64, period int) []float64 {
	n := len(data)
	result := make([]float64, n)

	// 预计算常数（只计算CTI需要的部分）
	periodF := float64(period)
	sumX := periodF * float64(period+1) * 0.5
	sumX2 := sumX * (2*periodF + 1) / 3
	divisor := periodF*sumX2 - sumX*sumX

	arr := make([]float64, 0, n)
	// 滑动窗口计算，专门为相关系数r优化
	for i := 0; i < n; i++ {
		val := data[i]
		if math.IsNaN(val) {
			result[i] = math.NaN()
			continue
		}
		arr = append(arr, val)
		if len(arr) > period {
			arr = arr[1:]
		} else if len(arr) < period {
			result[i] = math.NaN()
			continue
		}
		// 检查窗口中是否有NaN值，同时计算所需的和
		var sumY, sumXY, sumY2 float64

		for j, v := range arr {
			sumY += v
			sumXY += float64(j+1) * v
			sumY2 += v * v
		}

		// 计算相关系数 r
		rn := periodF*sumXY - sumX*sumY
		rd := math.Sqrt(divisor * (periodF*sumY2 - sumY*sumY))
		result[i] = rn / rd
	}

	return result
}

// Highest finds the highest value over a preceding period for each point.
// No changes needed as it relies on the correct slidingWindow implementation.
func Highest(data []float64, period int) []float64 {
	return slidingWindow(data, period, true)
}

// HighestBar finds the offset of the highest bar in a preceding period.
func HighestBar(data []float64, period int) []float64 {
	return findExtremeBarOffset(data, period, false)
}

// LINEARREG_ANGLE is the conventional TA-Lib spelling.
func LINEARREG_ANGLE(data []float64, period int) []float64 {
	return LinearRegAngle(data, period)
}

// LinRegAdv 计算线性回归的各种指标，并行版本
func LinReg(data []float64, period int) []float64 {
	return LinRegAdv(data, period, false, false, false, false, false, false)
}

// LinRegAdv 计算线性回归的各种指标，并行版本
func LinRegAdv(data []float64, period int, angle, intercept, degrees, r, slope, tsf bool) []float64 {
	n := len(data)
	result := make([]float64, n)

	// 预计算常数
	sumX, sumX2, periodF := precomputeLinRegConstants(period)

	// 滑动窗口计算
	arr := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		// 获取当前窗口的数据
		val := data[i]
		if math.IsNaN(val) {
			result[i] = math.NaN()
			continue
		}
		arr = append(arr, val)
		if len(arr) > period {
			arr = arr[1:]
		} else if len(arr) < period {
			result[i] = math.NaN()
			continue
		}

		// 计算线性回归系数
		m, b, rVal := computeLinRegCoeff(arr, period, sumX, sumX2, periodF)

		// 根据需要的输出类型返回相应的值
		if slope {
			result[i] = m
		} else if intercept {
			result[i] = b
		} else if angle {
			theta := math.Atan(m)
			if degrees {
				theta *= 180 / math.Pi
			}
			result[i] = theta
		} else if r {
			result[i] = rVal
		} else if tsf {
			result[i] = m*periodF + b
		} else {
			result[i] = m*(periodF-1) + b
		}
	}

	return result
}

// LinearRegAngle returns TA-Lib LINEARREG_ANGLE values in degrees.
func LinearRegAngle(data []float64, period int) []float64 {
	return LinRegAdv(data, period, true, false, true, false, false, false)
}

// Lowest finds the lowest value over a preceding period for each point.
// No changes needed as it relies on the correct slidingWindow implementation.
func Lowest(data []float64, period int) []float64 {
	return slidingWindow(data, period, false)
}

// LowestBar finds the offset of the lowest bar in a preceding period.
func LowestBar(data []float64, period int) []float64 {
	return findExtremeBarOffset(data, period, true)
}

// Slope is the least-squares linear-regression slope over period bars.
func Slope(data []float64, period int) []float64 {
	return LinRegAdv(data, period, false, false, false, false, true, false)
}

func Correlation(a, b []float64, period int) []float64 {
	o := make([]float64, len(a))
	for i := range o {
		o[i] = math.NaN()
	}
	for i := period - 1; i < len(a) && period > 0; i++ {
		var sx, sy, sxx, syy, sxy float64
		for j := i - period + 1; j <= i; j++ {
			sx += a[j]
			sy += b[j]
			sxx += a[j] * a[j]
			syy += b[j] * b[j]
			sxy += a[j] * b[j]
		}
		n := float64(period)
		den := math.Sqrt((n*sxx - sx*sx) * (n*syy - sy*sy))
		if den != 0 {
			o[i] = (n*sxy - sx*sy) / den
		}
	}
	return o
}
