package banta

import "github.com/banbox/banta/tav"
import "math"

// Donchian returns upper, middle and lower channels, including the current
// bar in the rolling window. The output order is upper, middle, lower.
func Donchian(high, low *Series, period int) (*Series, *Series, *Series) {
	res := high.To("_donchian", period)
	if !res.Cached() {
		up, dn := math.NaN(), math.NaN()
		if period > 0 && high.Len() >= period && !math.IsNaN(high.Get(0)) && !math.IsNaN(low.Get(0)) {
			up, dn = high.Get(0), low.Get(0)
			for i := 1; i < period; i++ {
				if high.Get(i) > up {
					up = high.Get(i)
				}
				if low.Get(i) < dn {
					dn = low.Get(i)
				}
			}
		}
		if math.IsNaN(up) || math.IsNaN(dn) {
			res.Append([]float64{math.NaN(), math.NaN(), math.NaN()})
		} else {
			res.Append([]float64{up, (up + dn) / 2, dn})
		}
	}
	return res, res.Cols[0], res.Cols[1]
}

func DonchianPBand(high, low, close *Series, period int) *Series {
	r := close.To("_donchian_pband", period)
	if !r.Cached() {
		r.Append(last(tav.DonchianPBand(hist(high), hist(low), hist(close), period)))
	}
	return r
}

func KeltnerChannel(high, low, close *Series, period int, multiplier float64) (*Series, *Series, *Series) {
	k := pkey(period, multiplier)
	u := close.To("_kc_u", k)
	m := close.To("_kc_m", k)
	l := close.To("_kc_l", k)
	if u.Len() < close.Len() {
		a, b, c := tav.KeltnerChannel(hist(high), hist(low), hist(close), period, multiplier)
		u.Append(last(a))
		m.Append(last(b))
		l.Append(last(c))
	}
	return u, m, l
}

func KeltnerWBand(high, low, close *Series, period int, mult float64) *Series {
	r := close.To("_keltner_wband", period*1000+int(mult*100))
	if !r.Cached() {
		r.Append(last(tav.KeltnerWBand(hist(high), hist(low), hist(close), period, mult)))
	}
	return r
}

func PMAX(high, low, close *Series, period int, multiplier float64) (*Series, *Series) {
	k := pkey(period, multiplier)
	p := close.To("_pmax", k)
	d := close.To("_pmax_dir", k)
	if p.Len() < close.Len() {
		a, b := tav.PMAX(hist(high), hist(low), hist(close), period, multiplier)
		p.Append(last(a))
		d.Append(last(b))
	}
	return p, d
}
