package banta

import "github.com/banbox/banta/tav"
import "math"

/*
CMF Chaikin Money Flow

https://www.tradingview.com/scripts/chaikinmoneyflow/?solution=43000501974

suggest period: 20
*/
func CMF(env *BarEnv, period int) *Series {
	res := env.Close.To("_cmf", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		multiplier, volume := moneyFlowVol(env)
		mfVolume := multiplier * volume

		sta, _ := res.More.(*cmfState)
		if sta == nil {
			sta = &cmfState{}
			res.More = sta
			res.DupMore = func(more interface{}) interface{} {
				m := more.(*cmfState)
				return &cmfState{append([]float64{}, m.mfSum...), append([]float64{}, m.volSum...), m.sumMfVal, m.sumVol}
			}
		}

		var resVal = math.NaN()
		if math.IsNaN(mfVolume) || math.IsNaN(volume) || volume <= 0 {
		} else {
			// Sum the Money Flow Volumes and Volumes over the period
			sta.sumMfVal += mfVolume
			sta.sumVol += volume

			if len(sta.mfSum) < period {
				sta.mfSum = append(sta.mfSum, mfVolume)
				sta.volSum = append(sta.volSum, volume)
				if len(sta.mfSum) == period {
					resVal = sta.sumMfVal / sta.sumVol
				}
			} else {
				sta.sumMfVal -= sta.mfSum[0]
				sta.mfSum = append(sta.mfSum[1:], mfVolume)

				sta.sumVol -= sta.volSum[0]
				sta.volSum = append(sta.volSum[1:], volume)
				if sta.sumVol > 0 {
					// Calculate CMF = Sum(Money Flow Volume) / Sum(Volume)
					resVal = sta.sumMfVal / sta.sumVol
				}
			}
		}
		res.Append(resVal)
	}
	return res
}

/*
ChaikinOsc Chaikin Oscillator

https://www.tradingview.com/support/solutions/43000501979-chaikin-oscillator/

short: 3, long: 10
*/
func ChaikinOsc(env *BarEnv, shortLen int, longLen int) *Series {
	res := env.Close.To("_chaikinosc", shortLen*1000+longLen)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		adl := ADL(env)

		shortEma := EMA(adl, shortLen)
		longEma := EMA(adl, longLen)

		oscValue := shortEma.Get(0) - longEma.Get(0)
		res.Append(oscValue)
	}
	return res
}

// ADL Accumulation/Distribution Line
func ADL(env *BarEnv) *Series {
	adl := env.Close.To("_adl", 0)
	if adl.Cached() {
		return adl
	}
	if !adl.Cached() {
		multiplier, volume := moneyFlowVol(env)
		mfVolume := multiplier * volume

		adlValue := mfVolume
		if adl.Len() > 0 {
			adlValue += adl.Get(0)
		}
		adl.Append(adlValue)
	}
	return adl
}

// ADOSC is the Chaikin/Accumulation-Distribution oscillator. It returns
// EMA(ADL, fast) - EMA(ADL, slow), with fast and slow periods in bars.
func ADOSC(env *BarEnv, fast, slow int) *Series {
	key := fast*100000 + slow
	res := env.Close.To("_adosc", key)
	if res.Cached() {
		return res
	}
	if fast < 1 || slow < 1 {
		res.Append(math.NaN())
		return res
	}
	type state struct {
		adl                  []float64
		cum                  float64
		fast, slow           float64
		fastReady, slowReady bool
	}
	s, _ := res.More.(*state)
	if s == nil {
		s = &state{}
		res.More = s
		res.DupMore = func(v interface{}) interface{} {
			x := v.(*state)
			return &state{adl: append([]float64(nil), x.adl...), cum: x.cum, fast: x.fast, slow: x.slow, fastReady: x.fastReady, slowReady: x.slowReady}
		}
	}
	mult, vol := moneyFlowVol(env)
	s.cum += mult * vol
	s.adl = append(s.adl, s.cum)
	update := func(period int, previous *float64, ready *bool) float64 {
		if !*ready {
			*previous = s.adl[0]
			*ready = true
			return *previous
		}
		alpha := 2.0 / float64(period+1)
		*previous = alpha*s.adl[len(s.adl)-1] + (1-alpha)*(*previous)
		return *previous
	}
	f, sl := update(fast, &s.fast, &s.fastReady), update(slow, &s.slow, &s.slowReady)
	if len(s.adl) < slow || math.IsNaN(f) || math.IsNaN(sl) {
		res.Append(math.NaN())
	} else {
		res.Append(f - sl)
	}
	return res
}

// EFI is Elder's Force Index. It is EMA(period) of volume multiplied by the
// close-to-close change. The first bar has no change and is undefined.
func EFI(env *BarEnv, period int) *Series {
	key := period
	res := env.Close.To("_efi", key)
	if res.Cached() {
		return res
	}
	force := env.Close.To("_efi_force", key)
	prev := env.Close.Get(1)
	cur := env.Close.Get(0)
	if period < 1 || math.IsNaN(prev) || math.IsNaN(cur) {
		force.Append(math.NaN())
	} else {
		force.Append((cur - prev) * env.Volume.Get(0))
	}
	res.Append(EMA(force, period).Get(0))
	return res
}

// OBV is On Balance Volume (TA-Lib OBV).
func OBV(close, volume *Series) *Series {
	res := close.To("_obv", 0)
	if !res.Cached() {
		prev := close.Get(1)
		v := volume.Get(0)
		if res.Len() == 0 {
			res.Append(v)
		} else if close.Get(0) > prev {
			res.Append(res.Get(0) + v)
		} else if close.Get(0) < prev {
			res.Append(res.Get(0) - v)
		} else {
			res.Append(res.Get(0))
		}
	}
	return res
}

func VPCI(close, volume *Series, period int) *Series {
	r := close.To("_vpci", period)
	if !r.Cached() {
		r.Append(last(tav.VPCI(hist(close), hist(volume), period)))
	}
	return r
}

func moneyFlowVol(env *BarEnv) (float64, float64) {
	// Retrieve the latest values
	closeVal := env.Close.Get(0)
	high := env.High.Get(0)
	low := env.Low.Get(0)
	volume := env.Volume.Get(0)

	var multiplier float64

	// Money Flow Multiplier = [(Close - Low) - (High - Close)] / (High - Low)
	if high > low {
		multiplier = ((closeVal - low) - (high - closeVal)) / (high - low)
	}

	// Money Flow Volume = Money Flow Multiplier x Volume
	return multiplier, volume
}

/*
MFI Money Flow Index

https://corporatefinanceinstitute.com/resources/career-map/sell-side/capital-markets/money-flow-index/

suggest period: 14
*/
func MFI(e *BarEnv, period int) *Series {
	res := e.Close.To("_mfi", period)
	if res.Cached() {
		return res
	}
	if !res.Cached() {
		sta, _ := res.More.(*mfiState)
		if sta == nil {
			sta = &mfiState{sumNeg: math.NaN(), sumPos: math.NaN(), prev: math.NaN()}
			res.More = sta
			res.DupMore = func(more interface{}) interface{} {
				m := more.(*mfiState)
				return &mfiState{append([]float64{}, m.posArr...), append([]float64{}, m.negArr...), m.sumPos, m.sumNeg, m.prev}
			}
		}
		avgPrice := AvgPrice(e)
		price0 := avgPrice.Get(0)
		if math.IsNaN(price0) {
			res.Append(math.NaN())
		} else {
			moneyFlow := price0 * e.Volume.Get(0)
			posFlow, negFlow := float64(0), float64(0)
			if price0 > sta.prev {
				posFlow = moneyFlow
			} else if price0 < sta.prev {
				negFlow = moneyFlow
			}
			sta.prev = price0
			if math.IsNaN(moneyFlow) {
				res.Append(math.NaN())
			} else {
				var resVal = math.NaN()
				if math.IsNaN(sta.sumPos) || math.IsNaN(sta.sumNeg) {
					sta.posArr = []float64{posFlow}
					sta.negArr = []float64{negFlow}
					sta.sumPos = posFlow
					sta.sumNeg = negFlow
				} else {
					sta.sumPos += posFlow
					sta.sumNeg += negFlow
					sta.posArr = append(sta.posArr, posFlow)
					sta.negArr = append(sta.negArr, negFlow)
					if len(sta.posArr) >= period {
						if len(sta.posArr) > period {
							sta.sumPos -= sta.posArr[0]
							sta.sumNeg -= sta.negArr[0]
							sta.posArr = sta.posArr[1:]
							sta.negArr = sta.negArr[1:]
						}
						if sta.sumNeg > 0 {
							moneyFlowRatio := sta.sumPos / sta.sumNeg
							resVal = 100 - (100 / (1 + moneyFlowRatio))
						}
					}
				}
				res.Append(resVal)
			}
		}
	}
	return res
}
