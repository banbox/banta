package banta

import (
	"testing"

	"github.com/banbox/banta/tav"
)

func TestExtraStateVectorParity(t *testing.T) {
	ks := extraFixture()
	o, h, l, c, v, _ := extractOHLCV(ks)
	assertStateVectorParity(t, "MOM", tav.MOM(c, 2), func(e *BarEnv) float64 { return MOM(e.Close, 2).Get(0) })
	assertStateVectorParity(t, "OBV", tav.OBV(c, v), func(e *BarEnv) float64 { return OBV(e.Close, e.Volume).Get(0) })
	assertStateVectorParity(t, "DEMA", tav.DEMA(c, 3), func(e *BarEnv) float64 { return DEMA(e.Close, 3).Get(0) })
	assertStateVectorParity(t, "T3", tav.T3(c, 3), func(e *BarEnv) float64 { return T3(e.Close, 3).Get(0) })
	assertStateVectorParity(t, "AroonOsc", tav.AroonOsc(h, l, 3), func(e *BarEnv) float64 {
		return AroonOsc(e.High, e.Low, 3).Get(0)
	})
	assertStateVectorParity(t, "PivotHigh", tav.PivotHigh(h, 1, 1), func(e *BarEnv) float64 { return PivotHigh(e.High, 1, 1).Get(0) })
	assertStateVectorParity(t, "PivotLow", tav.PivotLow(l, 1, 1), func(e *BarEnv) float64 { return PivotLow(e.Low, 1, 1).Get(0) })
	stochK, _ := tav.StochF(h, l, c, 3)
	assertStateVectorParity(t, "StochF", stochK, func(e *BarEnv) float64 { k, _ := StochF(e.High, e.Low, e.Close, 3); return k.Get(0) })
	assertStateVectorParity(t, "ULTOSC", tav.ULTOSC(h, l, c, 2, 3, 4), func(e *BarEnv) float64 { return ULTOSC(e.High, e.Low, e.Close, 2, 3, 4).Get(0) })
	assertStateVectorParity(t, "SAR", tav.SAR(h, l, .02, .2), func(e *BarEnv) float64 { return SAR(e.High, e.Low, .02, .2).Get(0) })
	_ = o
}
