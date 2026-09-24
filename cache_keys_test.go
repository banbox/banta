package banta

import "testing"

func TestStatefulIndicatorCacheKeysIncludeAllParameters(t *testing.T) {
	e, err := NewBarEnv("test", "spot", "BTC/USDT", "1d")
	if err != nil {
		t.Fatal(err)
	}
	if err := e.OnBar(1, 99, 101, 98, 100, 10, 1000, 5, 1); err != nil {
		t.Fatal(err)
	}

	assertDistinct := func(name string, first, second *Series) {
		t.Helper()
		if first == second {
			t.Fatalf("%s reused a cached series for different parameters", name)
		}
	}

	macd1, _ := MACD(e.Close, 12, 26, 9)
	macd2, _ := MACD(e.Close, 12, 25, 19)
	assertDistinct("MACD", macd1, macd2)

	stc1 := STC(e.Close, 12, 26, 50, 0.1)
	stc2 := STC(e.Close, 12, 26, 50, 0.2)
	assertDistinct("STC", stc1, stc2)

	alma1 := ALMA(e.Close, 5, 1.00, 0.02)
	alma2 := ALMA(e.Close, 5, 0.99, 0.03)
	assertDistinct("ALMA", alma1, alma2)

	rma1 := RMABy(e.Close, 3, 0, 1.1)
	rma2 := RMABy(e.Close, 3, 0, 1.9)
	assertDistinct("RMABy", rma1, rma2)

	bb1, _, _ := BBANDS(e.Close, 5, 2.0, 0.11)
	bb2, _, _ := BBANDS(e.Close, 5, 2.0, 0.19)
	assertDistinct("BBANDS", bb1, bb2)

	atr := ATR(e.High, e.Low, e.Close, 3)
	ut1 := UTBot(e.Close, atr, 0.11)
	ut2 := UTBot(e.Close, atr, 0.19)
	assertDistinct("UTBot", ut1, ut2)

	sar1 := SAR(e.High, e.Low, 0.0101, 0.2)
	sar2 := SAR(e.High, e.Low, 0.0109, 0.2)
	assertDistinct("SAR", sar1, sar2)

	mama1, _ := MAMA(e.Close, 1.001, 2.01)
	mama2, _ := MAMA(e.Close, 1.002, 1.91)
	assertDistinct("MAMA", mama1, mama2)

	super1 := Supertrend(e.High, e.Low, e.Close, 5, 1.0011)
	super2 := Supertrend(e.High, e.Low, e.Close, 5, 1.0019)
	assertDistinct("Supertrend", super1, super2)

	keltner1 := KeltnerWBand(e.High, e.Low, e.Close, 5, 1.111)
	keltner2 := KeltnerWBand(e.High, e.Low, e.Close, 5, 1.119)
	assertDistinct("KeltnerWBand", keltner1, keltner2)

	stiff1 := Stiffness(e.Close, 5, 3, 2)
	stiff2 := Stiffness(e.Close, 5, 3, 4)
	assertDistinct("Stiffness", stiff1, stiff2)
}
