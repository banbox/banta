package tav

import "math"

// Donchian returns upper, middle and lower channels, including the current bar.
func Donchian(high, low []float64, period int) ([]float64, []float64, []float64) {
	u, m, d := make([]float64, len(high)), make([]float64, len(high)), make([]float64, len(high))
	for i := range high {
		u[i], m[i], d[i] = math.NaN(), math.NaN(), math.NaN()
	}
	for i := period - 1; period > 0 && i < len(high); i++ {
		up, dn := high[i], low[i]
		valid := !math.IsNaN(up) && !math.IsNaN(dn)
		for j := 1; j < period && valid; j++ {
			if math.IsNaN(high[i-j]) || math.IsNaN(low[i-j]) {
				valid = false
				break
			}
			if high[i-j] > up {
				up = high[i-j]
			}
			if low[i-j] < dn {
				dn = low[i-j]
			}
		}
		if valid {
			u[i], d[i], m[i] = up, dn, (up+dn)/2
		}
	}
	return u, m, d
}

// KeltnerChannel returns upper, middle EMA and lower bands.
func KeltnerChannel(high, low, close []float64, period int, multiplier float64) ([]float64, []float64, []float64) {
	m := EMA(close, period)
	a := ATR(high, low, close, period)
	u := make([]float64, len(close))
	l := make([]float64, len(close))
	for i := range close {
		u[i] = math.NaN()
		l[i] = math.NaN()
		if !math.IsNaN(m[i]) && !math.IsNaN(a[i]) {
			u[i] = m[i] + multiplier*a[i]
			l[i] = m[i] - multiplier*a[i]
		}
	}
	return u, m, l
}

// PMAX returns the profit-maximizer trailing line and direction (1/-1).
func PMAX(high, low, close []float64, period int, multiplier float64) ([]float64, []float64) {
	n := len(close)
	line := make([]float64, n)
	dir := make([]float64, n)
	for i := 0; i < n; i++ {
		line[i] = math.NaN()
		dir[i] = math.NaN()
	}
	ma := EMA(close, period)
	a := ATR(high, low, close, period)
	fu, fl := math.NaN(), math.NaN()
	up := true
	for i := 0; i < n; i++ {
		if math.IsNaN(ma[i]) || math.IsNaN(a[i]) {
			continue
		}
		bu, bl := ma[i]+multiplier*a[i], ma[i]-multiplier*a[i]
		if math.IsNaN(fu) {
			fu, fl = bu, bl
		} else {
			if bu < fu || close[i-1] > fu {
				fu = bu
			}
			if bl > fl || close[i-1] < fl {
				fl = bl
			}
			if up && ma[i] < fl {
				up = false
			} else if !up && ma[i] > fu {
				up = true
			}
		}
		if up {
			line[i] = fl
			dir[i] = 1
		} else {
			line[i] = fu
			dir[i] = -1
		}
	}
	return line, dir
}

func DonchianPBand(high, low, close []float64, period int) []float64 {
	u, m, d := Donchian(high, low, period)
	_ = m
	o := make([]float64, len(close))
	for i := range o {
		o[i] = math.NaN()
		if !math.IsNaN(u[i]) && !math.IsNaN(d[i]) && u[i] != d[i] {
			o[i] = (close[i] - d[i]) / (u[i] - d[i])
		}
	}
	return o
}

func KeltnerWBand(high, low, close []float64, period int, mult float64) []float64 {
	u, m, d := KeltnerChannel(high, low, close, period, mult)
	o := make([]float64, len(close))
	for i := range o {
		o[i] = math.NaN()
		if !math.IsNaN(m[i]) && m[i] != 0 {
			o[i] = (u[i] - d[i]) / m[i]
		}
	}
	return o
}
