package banta

import "math"

// PivotHigh/PivotLow confirm an extremum left/right bars after it occurs,
// matching Pine ta.pivothigh/ta.pivotlow output timing.
func PivotHigh(src *Series, left, right int) *Series {
	res := src.To("_pivothigh", left*1000+right)
	if !res.Cached() {
		idx := right
		if len(src.Data) < left+right+1 {
			res.Append(math.NaN())
		} else {
			v := src.Get(idx)
			ok := true
			for i := idx - right; i <= idx+left; i++ {
				if i != idx && src.Get(i) >= v {
					ok = false
					break
				}
			}
			if ok {
				res.Append(v)
			} else {
				res.Append(math.NaN())
			}
		}
	}
	return res
}

func CDL3INSIDE(open, high, low, close *Series) *Series {
	res := close.To("_cdl3inside", 0)
	if !res.Cached() {
		if close.Len() < 3 {
			res.Append(0.0)
		} else {
			o0, c0, o1, c1, o2, c2 := open.Get(0), close.Get(0), open.Get(1), close.Get(1), open.Get(2), close.Get(2)
			insideBull := c2 < o2 && c1 > o1 && o1 > c2 && c1 < o2 && c0 > o0 && c0 > c1
			insideBear := c2 > o2 && c1 < o1 && o1 < c2 && c1 > o2 && c0 < o0 && c0 < c1
			if insideBull {
				res.Append(100.0)
			} else if insideBear {
				res.Append(-100.0)
			} else {
				res.Append(0.0)
			}
		}
	}
	return res
}

func CDL3LINESTRIKE(open, high, low, close *Series) *Series {
	res := close.To("_cdl3linestrike", 0)
	if !res.Cached() {
		if close.Len() < 4 {
			res.Append(0.0)
		} else {
			o0, c0, o1, c1, o2, c2, o3, c3 := open.Get(0), close.Get(0), open.Get(1), close.Get(1), open.Get(2), close.Get(2), open.Get(3), close.Get(3)
			bull := c3 < o3 && c2 < o2 && c1 < o1 && c0 > o0 && o0 < c3 && c0 > o3
			bear := c3 > o3 && c2 > o2 && c1 > o1 && c0 < o0 && o0 > c3 && c0 < o3
			if bull {
				res.Append(100.0)
			} else if bear {
				res.Append(-100.0)
			} else {
				res.Append(0.0)
			}
		}
	}
	return res
}

func CDL3OUTSIDE(open, high, low, close *Series) *Series {
	res := close.To("_cdl3outside", 0)
	if !res.Cached() {
		if close.Len() < 3 {
			res.Append(0.0)
		} else {
			c0, o1, c1, o2, c2 := close.Get(0), open.Get(1), close.Get(1), open.Get(2), close.Get(2)
			bull := c2 < o2 && c1 > o1 && o1 <= c2 && c1 >= o2 && c0 > c1
			bear := c2 > o2 && c1 < o1 && o1 >= c2 && c1 <= o2 && c0 < c1
			if bull {
				res.Append(100.0)
			} else if bear {
				res.Append(-100.0)
			} else {
				res.Append(0.0)
			}
		}
	}
	return res
}

func CDLDRAGONFLYDOJI(open, high, low, close *Series) *Series {
	return candlePatternSingle(open, high, low, close, "_cdldragonfly", func(o, h, l, c float64) bool {
		b, u, d := candleParts(o, h, l, c)
		return b <= (h-l)*0.1 && u <= (h-l)*0.1 && d >= (h-l)*0.6
	})
}

func CDLENGULFING(open, high, low, close *Series) *Series {
	res := close.To("_cdlengulfing", 0)
	if !res.Cached() {
		if close.Len() < 2 {
			res.Append(0.0)
		} else {
			o, c, po, pc := open.Get(0), close.Get(0), open.Get(1), close.Get(1)
			bull := pc < po && c > o && o <= pc && c >= po
			bear := pc > po && c < o && o >= pc && c <= po
			if bull {
				res.Append(100.0)
			} else if bear {
				res.Append(-100.0)
			} else {
				res.Append(0.0)
			}
		}
	}
	return res
}

func CDLGRAVESTONEDOJI(open, high, low, close *Series) *Series {
	return candlePatternSingle(open, high, low, close, "_cdlgravestone", func(o, h, l, c float64) bool {
		b, u, d := candleParts(o, h, l, c)
		return b <= (h-l)*0.1 && d <= (h-l)*0.1 && u >= (h-l)*0.6
	})
}

func CDLHAMMER(open, high, low, close *Series) *Series {
	res := close.To("_cdlhammer", 0)
	if !res.Cached() {
		if close.Len() == 0 {
			res.Append(math.NaN())
		} else {
			b, u, d := candleParts(open.Get(0), high.Get(0), low.Get(0), close.Get(0))
			if d >= 2*b && u <= b && math.Max(open.Get(0), close.Get(0)) >= high.Get(0)-0.25*(high.Get(0)-low.Get(0)) {
				res.Append(100.0)
			} else {
				res.Append(0.0)
			}
		}
	}
	return res
}

func CDLHANGINGMAN(open, high, low, close *Series) *Series {
	return candlePatternSingle(open, high, low, close, "_cdlhangingman", func(o, h, l, c float64) bool { b, u, d := candleParts(o, h, l, c); return d >= 2*b && u <= b && c < o })
}

func CDLMORNINGSTAR(open, high, low, close *Series) *Series {
	return candlePatternSingle(open, high, low, close, "_cdlmorningstar", func(o, h, l, c float64) bool {
		return c > o && close.Get(1) < open.Get(1) && close.Get(2) < open.Get(2)
	})
}

func CDLSHOOTINGSTAR(open, high, low, close *Series) *Series {
	return candlePatternSingle(open, high, low, close, "_cdlshootingstar", func(o, h, l, c float64) bool { b, u, d := candleParts(o, h, l, c); return u >= 2*b && d <= b && c < o })
}

func PivotLow(src *Series, left, right int) *Series {
	res := src.To("_pivotlow", left*1000+right)
	if !res.Cached() {
		idx := right
		if len(src.Data) < left+right+1 {
			res.Append(math.NaN())
		} else {
			v := src.Get(idx)
			ok := true
			for i := idx - right; i <= idx+left; i++ {
				if i != idx && src.Get(i) <= v {
					ok = false
					break
				}
			}
			if ok {
				res.Append(v)
			} else {
				res.Append(math.NaN())
			}
		}
	}
	return res
}

func candleParts(o, h, l, c float64) (body, upper, lower float64) {
	body = math.Abs(c - o)
	upper = h - math.Max(o, c)
	lower = math.Min(o, c) - l
	return
}

func candlePatternSingle(open, high, low, close *Series, key string, match func(float64, float64, float64, float64) bool) *Series {
	res := close.To(key, 0)
	if !res.Cached() {
		if match(open.Get(0), high.Get(0), low.Get(0), close.Get(0)) {
			res.Append(100.0)
		} else {
			res.Append(0.0)
		}
	}
	return res
}

// HeikinAshi return [open,high,low,close]
func HeikinAshi(e *BarEnv) (*Series, *Series, *Series, *Series) {
	res := e.Close.To("_heikin", 0)
	if !res.Cached() {
		if !res.Cached() {
			ho := e.Open.To("_hka", 0)
			hh := e.High.To("_hka", 0)
			hl := e.Low.To("_hka", 0)
			hc := e.Close.To("_hka", 0)

			o, h, l, c := e.Open.Get(0), e.High.Get(0), e.Low.Get(0), e.Close.Get(0)

			po := ho.Get(0)
			if math.IsNaN(po) {
				ho.Append((o + c) / 2)
			} else {
				ho.Append((po + hc.Get(0)) / 2)
			}
			hcVal := (o + h + l + c) / 4
			hc.Append(hcVal)
			hoVal := ho.Get(0)
			hh.Append(max(h, hoVal, hcVal))
			hl.Append(min(l, hoVal, hcVal))

			res.Append([]*Series{ho, hh, hl, hc})
		}
	}

	return res, res.Cols[0], res.Cols[1], res.Cols[2]
}
