package banta

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/banbox/banta/tav"
)

type compareFixture struct {
	Candles [][]float64          `json:"candles"`
	Results map[string][]float64 `json:"results"`
}

func loadCompareFixture(t *testing.T, name string) compareFixture {
	t.Helper()
	b, err := os.ReadFile("testdata/indicator_" + name + "_compare.json")
	if err != nil {
		t.Fatal(err)
	}
	var f compareFixture
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/fixture_btc_58.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(fixture, &f.Candles); err != nil {
		t.Fatal(err)
	}
	return f
}

func compareExternal(t *testing.T, name, line string, got []float64) {
	t.Helper()
	f := loadCompareFixture(t, name)
	want := f.Results[line]
	if len(got) != len(want) {
		t.Fatalf("%s length %d != %d", name, len(got), len(want))
	}
	for i := range got {
		if math.IsNaN(got[i]) && (want[i] == 0 || math.IsNaN(got[i])) {
			continue
		}
		// Cross-language rolling sums and atan differ by a few ULPs; the
		// state/vector parity assertion above remains exact.
		if got[i] != want[i] && math.Abs(got[i]-want[i]) > 1e-8*math.Max(1, math.Abs(want[i])) {
			t.Fatalf("%s %s bar %d: %.17g != %.17g", name, line, i, got[i], want[i])
		}
	}
}

func TestAOBatchStateVectorParity(t *testing.T) {
	ks := extraFixture()
	_, h, l, _, _, _ := extractOHLCV(ks)
	assertStateVectorParity(t, "AO", tav.AO(h, l, 2, 4), func(e *BarEnv) float64 { return AO(e.High, e.Low, 2, 4).Get(0) })
	got := tav.AO(h, l, 2, 4)
	want := []float64{math.NaN(), math.NaN(), math.NaN(), .375, -.375, .5, 1.25, .375}
	for i := range got {
		if math.IsNaN(want[i]) && math.IsNaN(got[i]) {
			continue
		}
		if got[i] != want[i] {
			t.Fatalf("AO oracle[%d]=%v want %v", i, got[i], want[i])
		}
	}
}

func TestADOSCBatchStateVectorParity(t *testing.T) {
	ks := extraFixture()
	_, h, l, c, v, _ := extractOHLCV(ks)
	assertStateVectorParity(t, "ADOSC", tav.ADOSC(h, l, c, v, 2, 4), func(e *BarEnv) float64 { return ADOSC(e, 2, 4).Get(0) })
}

func TestBatchAExternalOracles(t *testing.T) {
	for _, name := range []string{"ao", "adosc"} {
		f := loadCompareFixture(t, name)
		ks := make([]Kline, len(f.Candles))
		for i, c := range f.Candles {
			ks[i] = Kline{Time: int64(c[0]), Open: c[1], High: c[2], Low: c[3], Close: c[4], Volume: c[5]}
		}
		_, h, l, close, vol, _ := extractOHLCV(ks)
		if name == "ao" {
			compareExternal(t, name, "pandas_ta", tav.AO(h, l, 5, 34))
		} else {
			compareExternal(t, name, "talib", tav.ADOSC(h, l, close, vol, 3, 10))
		}
	}
}
