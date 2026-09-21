package banta

/*
AvgPrice typical price=(h+l+c)/3
*/
func AvgPrice(e *BarEnv) *Series {
	res := e.Close.To("_avgp", 0)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		avgPrice := (e.High.Get(0) + e.Low.Get(0) + e.Close.Get(0)) / 3
		res.Append(avgPrice)
	}
	return res
}

// HLC3 typical price=(h+l+c)/3
func HLC3(h, l, c *Series) *Series {
	res := c.To("_hlc3", 0)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		res.Append((h.Get(0) + l.Get(0) + c.Get(0)) / 3)
	}
	return res
}

// HLCC4 and OHLC4 are source-language price composites.
func HLCC4(h, l, c *Series) *Series { return h.Add(l).Add(c).Add(c).Div(4) }

func Abs(a *Series) *Series { return a.Abs() }

func Change(a *Series) *Series { return a.Sub(a.Back(1)) }

func HL2(h, l *Series) *Series {
	res := l.To("_hl", 0)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		avgPrice := (h.Get(0) + l.Get(0)) / 2
		res.Append(avgPrice)
	}
	return res
}

func OHLC4(o, h, l, c *Series) *Series { return o.Add(h).Add(l).Add(c).Div(4) }

func Sub(a, b *Series) *Series { return a.Sub(b) }
