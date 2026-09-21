package banta

import (
	"math"
	"testing"
)

func extraFixture() []Kline {
	return []Kline{
		{Time: 86400000, Open: 10, High: 12, Low: 9, Close: 11, Volume: 5},
		{Time: 172800000, Open: 11, High: 13, Low: 10, Close: 12, Volume: 7},
		{Time: 259200000, Open: 12, High: 14, Low: 11, Close: 13, Volume: 6},
		{Time: 345600000, Open: 13, High: 13, Low: 9, Close: 10, Volume: 8},
		{Time: 432000000, Open: 10, High: 15, Low: 8, Close: 14, Volume: 9},
		{Time: 518400000, Open: 14, High: 16, Low: 12, Close: 15, Volume: 4},
		{Time: 604800000, Open: 15, High: 16, Low: 11, Close: 12, Volume: 10},
		{Time: 691200000, Open: 12, High: 17, Low: 10, Close: 16, Volume: 11},
	}
}

func assertStateVectorParity(t *testing.T, name string, vector []float64, run func(*BarEnv) float64) {
	t.Helper()
	state := make([]float64, 0, len(extraFixture()))
	env, err := NewBarEnv("test", "spot", "", "1d")
	if err != nil {
		t.Fatal(err)
	}
	RunFakeEnv(env, extraFixture(), func(_ int, _ Kline) { state = append(state, run(env)) })
	assertFloatSliceParity(t, name, vector, state, 1e-9)
}

func assertFloatSliceParity(t *testing.T, name string, got, want []float64, tolerance float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length %d != %d", name, len(got), len(want))
	}
	for i := range got {
		if math.IsNaN(got[i]) && math.IsNaN(want[i]) {
			continue
		}
		if math.Abs(got[i]-want[i]) > tolerance {
			t.Fatalf("%s[%d]: %.17g != %.17g", name, i, got[i], want[i])
		}
	}
}
