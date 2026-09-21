package tav

import "math"

func CDL3INSIDE(o, h, l, c []float64) []float64 {
	r := make([]float64, len(c))
	for i := 2; i < len(c); i++ {
		bull := c[i-2] < o[i-2] && c[i-1] > o[i-1] && o[i-1] > c[i-2] && c[i-1] < o[i-2] && c[i] > o[i] && c[i] > c[i-1]
		bear := c[i-2] > o[i-2] && c[i-1] < o[i-1] && o[i-1] < c[i-2] && c[i-1] > o[i-2] && c[i] < o[i] && c[i] < c[i-1]
		if bull {
			r[i] = 100
		} else if bear {
			r[i] = -100
		}
	}
	return r
}

func CDL3LINESTRIKE(o, h, l, c []float64) []float64 {
	r := make([]float64, len(c))
	for i := 3; i < len(c); i++ {
		bull := c[i-3] < o[i-3] && c[i-2] < o[i-2] && c[i-1] < o[i-1] && c[i] > o[i] && o[i] < c[i-3] && c[i] > o[i-3]
		bear := c[i-3] > o[i-3] && c[i-2] > o[i-2] && c[i-1] > o[i-1] && c[i] < o[i] && o[i] > c[i-3] && c[i] < o[i-3]
		if bull {
			r[i] = 100
		} else if bear {
			r[i] = -100
		}
	}
	return r
}

func CDL3OUTSIDE(o, h, l, c []float64) []float64 {
	r := make([]float64, len(c))
	for i := 2; i < len(c); i++ {
		bull := c[i-2] < o[i-2] && c[i-1] > o[i-1] && o[i-1] <= c[i-2] && c[i-1] >= o[i-2] && c[i] > c[i-1]
		bear := c[i-2] > o[i-2] && c[i-1] < o[i-1] && o[i-1] >= c[i-2] && c[i-1] <= o[i-2] && c[i] < c[i-1]
		if bull {
			r[i] = 100
		} else if bear {
			r[i] = -100
		}
	}
	return r
}

func CDLDRAGONFLYDOJI(o, h, l, c []float64) []float64 {
	r := make([]float64, len(c))
	for i := range r {
		b, u, d := candleParts(o[i], h[i], l[i], c[i])
		if b <= (h[i]-l[i])*.1 && u <= (h[i]-l[i])*.1 && d >= (h[i]-l[i])*.6 {
			r[i] = 100
		}
	}
	return r
}

func CDLENGULFING(o, h, l, c []float64) []float64 {
	r := make([]float64, len(c))
	for i := 1; i < len(c); i++ {
		bull := c[i-1] < o[i-1] && c[i] > o[i] && o[i] <= c[i-1] && c[i] >= o[i-1]
		bear := c[i-1] > o[i-1] && c[i] < o[i] && o[i] >= c[i-1] && c[i] <= o[i-1]
		if bull {
			r[i] = 100
		} else if bear {
			r[i] = -100
		}
	}
	return r
}

func CDLGRAVESTONEDOJI(o, h, l, c []float64) []float64 {
	r := make([]float64, len(c))
	for i := range r {
		b, u, d := candleParts(o[i], h[i], l[i], c[i])
		if b <= (h[i]-l[i])*.1 && d <= (h[i]-l[i])*.1 && u >= (h[i]-l[i])*.6 {
			r[i] = 100
		}
	}
	return r
}

func CDLHAMMER(o, h, l, c []float64) []float64 {
	r := make([]float64, len(c))
	for i := range r {
		b, u, d := candleParts(o[i], h[i], l[i], c[i])
		if d >= 2*b && u <= b && math.Max(o[i], c[i]) >= h[i]-.25*(h[i]-l[i]) {
			r[i] = 100
		}
	}
	return r
}

func CDLHANGINGMAN(o, h, l, c []float64) []float64 {
	r := make([]float64, len(c))
	for i := range r {
		b, u, d := candleParts(o[i], h[i], l[i], c[i])
		if d >= 2*b && u <= b && c[i] < o[i] {
			r[i] = 100
		}
	}
	return r
}

func CDLMORNINGSTAR(o, h, l, c []float64) []float64 {
	r := make([]float64, len(c))
	for i := 2; i < len(c); i++ {
		if c[i] > o[i] && c[i-1] < o[i-1] && c[i-2] < o[i-2] {
			r[i] = 100
		}
	}
	return r
}

func CDLSHOOTINGSTAR(o, h, l, c []float64) []float64 {
	r := make([]float64, len(c))
	for i := range r {
		b, u, d := candleParts(o[i], h[i], l[i], c[i])
		if u >= 2*b && d <= b && c[i] < o[i] {
			r[i] = 100
		}
	}
	return r
}

func PivotHigh(data []float64, left, right int) []float64 {
	r := make([]float64, len(data))
	for i := range r {
		r[i] = math.NaN()
		idx := i - right
		if idx < left || idx+right >= len(data) {
			continue
		}
		v := data[idx]
		ok := true
		for j := idx - left; j <= idx+right; j++ {
			if j != idx && data[j] >= v {
				ok = false
				break
			}
		}
		if ok {
			r[i] = v
		}
	}
	return r
}

func PivotLow(data []float64, left, right int) []float64 {
	r := make([]float64, len(data))
	for i := range r {
		r[i] = math.NaN()
		idx := i - right
		if idx < left || idx+right >= len(data) {
			continue
		}
		v := data[idx]
		ok := true
		for j := idx - left; j <= idx+right; j++ {
			if j != idx && data[j] <= v {
				ok = false
				break
			}
		}
		if ok {
			r[i] = v
		}
	}
	return r
}

func candleParts(o, h, l, c float64) (body, upper, lower float64) {
	return math.Abs(c - o), h - math.Max(o, c), math.Min(o, c) - l
}

// HeikinAshi 并行版本，返回 [open, high, low, close] 数组
// 参数：开盘价、最高价、最低价、收盘价数组
func HeikinAshi(open, high, low, close []float64) ([]float64, []float64, []float64, []float64) {
	n := len(open)
	if n == 0 || len(high) != n || len(low) != n || len(close) != n {
		return nil, nil, nil, nil
	}

	hOpen := make([]float64, n)
	hHigh := make([]float64, n)
	hLow := make([]float64, n)
	hClose := make([]float64, n)

	pho := math.NaN()
	for i := 0; i < n; i++ {
		// Heikin Ashi Close 不依赖于前一个状态，可直接计算
		hClose[i] = (open[i] + high[i] + low[i] + close[i]) / 4

		if math.IsNaN(pho) {
			// 第一个有效蜡烛的开盘价
			hOpen[i] = (open[i] + close[i]) / 2
		} else {
			// 后续有效蜡烛的开盘价
			hOpen[i] = (hOpen[i-1] + hClose[i-1]) / 2
		}
		pho = hOpen[i]

		// Heikin Ashi High
		hHigh[i] = max(high[i], pho, hClose[i])

		// Heikin Ashi Low
		hLow[i] = min(low[i], pho, hClose[i])
	}

	return hOpen, hHigh, hLow, hClose
}
