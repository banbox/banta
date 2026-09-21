package tav

import "math"

// ADOSC is the Chaikin/Accumulation-Distribution oscillator.
func ADOSC(high, low, close, volume []float64, fast, slow int) []float64 {
	r := make([]float64, len(close))
	if fast < 1 || slow < 1 {
		for i := range r {
			r[i] = math.NaN()
		}
		return r
	}
	adl := make([]float64, len(close))
	var total float64
	for i := range close {
		mf := 0.0
		if high[i] > low[i] {
			mf = ((close[i] - low[i]) - (high[i] - close[i])) / (high[i] - low[i])
		}
		total += mf * volume[i]
		adl[i] = total
	}
	// TA-Lib ADOSC seeds both EMA streams with the first ADL value, unlike
	// Banta's public EMA which uses an initial SMA.
	f := EMABy(adl, fast, 1)
	s := EMABy(adl, slow, 1)
	for i := range r {
		if i+1 < slow {
			r[i] = math.NaN()
		} else {
			r[i] = f[i] - s[i]
		}
	}
	return r
}

// CMF Chaikin Money Flow
// period: 20
func CMF(high, low, close, volume []float64, period int) []float64 {
	n := len(close)
	res := make([]float64, n)

	// Not enough data to calculate
	if n < period {
		for i := range res {
			res[i] = math.NaN()
		}
		return res
	}

	// Pass 1: Calculate Money Flow Volume for each bar
	mfv := make([]float64, n)
	for i := 0; i < n; i++ {
		h, l, c := high[i], low[i], close[i]
		hilo := h - l

		if hilo > 0 {
			// Money Flow Multiplier = [(Close - Low) - (High - Close)] / (High - Low)
			multiplier := ((c - l) - (h - c)) / hilo
			// Money Flow Volume = Money Flow Multiplier x Volume
			mfv[i] = multiplier * volume[i]
		} else {
			mfv[i] = 0.0 // MFV is 0 if high equals low
		}
	}

	var mfvSum, volSum float64

	// Pass 2: Use a sliding window to calculate CMF
	okNum := 0
	for i := 0; i < n; i++ {
		mfvVal, vol := mfv[i], volume[i]
		if math.IsNaN(mfvVal) || math.IsNaN(vol) || vol <= 0 {
			res[i] = math.NaN()
			mfv = moveToFront(mfv, i)
			volume = moveToFront(volume, i)
			continue
		}
		mfvSum += mfvVal
		volSum += vol
		okNum += 1

		// Subtract the value that fell out of the window
		if okNum > period {
			mfvSum -= mfv[i-period]
			volSum -= volume[i-period]
		}

		if okNum >= period {
			// CMF = Sum(Money Flow Volume) / Sum(Volume)
			res[i] = mfvSum / volSum
		} else {
			// Fill initial values with NaN
			res[i] = math.NaN()
		}
	}
	return res
}

// EFI is EMA(period) of volume multiplied by the close-to-close change.
func EFI(close, volume []float64, period int) []float64 {
	r := make([]float64, len(close))
	force := make([]float64, len(close))
	for i := range r {
		r[i], force[i] = math.NaN(), math.NaN()
	}
	for i := 1; i < len(close); i++ {
		if !math.IsNaN(close[i]) && !math.IsNaN(close[i-1]) && !math.IsNaN(volume[i]) {
			force[i] = (close[i] - close[i-1]) * volume[i]
		}
	}
	if period < 1 {
		return r
	}
	return EMA(force, period)
}

// OBV computes On Balance Volume.
func OBV(close, volume []float64) []float64 {
	r := make([]float64, len(close))
	if len(close) == 0 {
		return r
	}
	r[0] = volume[0]
	for i := 1; i < len(close); i++ {
		r[i] = r[i-1]
		if close[i] > close[i-1] {
			r[i] += volume[i]
		} else if close[i] < close[i-1] {
			r[i] -= volume[i]
		}
	}
	return r
}

func VPCI(close, volume []float64, period int) []float64 {
	v := VWMA(close, volume, period)
	s := SMA(close, period)
	o := make([]float64, len(close))
	for i := range o {
		o[i] = math.NaN()
		if !math.IsNaN(v[i]) && !math.IsNaN(s[i]) {
			o[i] = v[i] - s[i]
		}
	}
	return o
}

// MFI 资金流量指数 - 并行计算版本
func MFI(high, low, close, volume []float64, period int) []float64 {
	n := len(close)
	res := make([]float64, n)

	// 计算正负资金流量
	posArr := make([]float64, 0, n)
	negArr := make([]float64, 0, n)

	// 计算滑动窗口的正负资金流量总和
	var sumPos, sumNeg float64

	prevPrice := math.NaN()
	for i := 0; i < n; i++ {
		cPrice := (high[i] + low[i] + close[i]) / 3.0
		if math.IsNaN(cPrice) {
			res[i] = math.NaN()
			continue
		}
		moneyFlow := cPrice * volume[i]
		var posFlow, negFlow float64
		if cPrice > prevPrice {
			posFlow = moneyFlow
		} else if cPrice < prevPrice {
			negFlow = moneyFlow
		}
		prevPrice = cPrice
		if math.IsNaN(moneyFlow) {
			res[i] = math.NaN()
			continue
		}
		posArr = append(posArr, posFlow)
		negArr = append(negArr, negFlow)
		sumPos += posFlow
		sumNeg += negFlow
		if len(posArr) >= period {
			if len(posArr) > period {
				// 滑动窗口：减去最旧的值
				sumPos -= posArr[0]
				sumNeg -= negArr[0]
				posArr = posArr[1:]
				negArr = negArr[1:]
			}
			// 计算MFI值
			if sumNeg > 0 {
				moneyFlowRatio := sumPos / sumNeg
				res[i] = 100 - (100 / (1 + moneyFlowRatio))
			} else {
				res[i] = math.NaN()
			}
		} else {
			res[i] = math.NaN()
		}
	}
	return res
}
