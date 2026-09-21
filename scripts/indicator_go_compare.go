package main

// Runs both Banta APIs over the shared fixture referenced by scripts/test_inds.py.
// The Python driver compares the JSON output with the selected external oracle.

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/banbox/banta"
	"github.com/banbox/banta/tav"
)

type input struct {
	Indicator string      `json:"indicator"`
	Fixture   string      `json:"fixture"`
	Candles   [][]float64 `json:"candles"`
}

type output struct {
	Indicator string                   `json:"indicator"`
	Vector    map[string][]interface{} `json:"vector"`
	State     map[string][]interface{} `json:"state"`
}

func clean(values []float64) []interface{} {
	result := make([]interface{}, len(values))
	for i, v := range values {
		if math.IsNaN(v) {
			result[i] = nil
		} else {
			result[i] = v
		}
	}
	return result
}

func main() {
	if len(os.Args) != 2 {
		panic("usage: indicator_go_compare <test_inds.json>")
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var in input
	if err := json.Unmarshal(b, &in); err != nil {
		panic(err)
	}
	if len(in.Candles) == 0 {
		fixture := in.Fixture
		if fixture == "" {
			fixture = "testdata/fixture_btc_58.json"
		} else if !strings.Contains(fixture, "/") {
			fixture = "testdata/" + fixture
		}
		fixtureBytes, err := os.ReadFile(fixture)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(fixtureBytes, &in.Candles); err != nil {
			panic(err)
		}
	}
	in.Indicator = canonical(in.Indicator)
	ks := make([]banta.Kline, len(in.Candles))
	for i, c := range in.Candles {
		if len(c) < 6 {
			panic("candle must contain time/open/high/low/close/volume")
		}
		ks[i] = banta.Kline{Time: int64(c[0]), Open: c[1], High: c[2], Low: c[3], Close: c[4], Volume: c[5]}
	}
	_, h, l, c, v, _ := extract(ks)
	vec := map[string][]float64{}
	stateRun := func(env *banta.BarEnv, name string) map[string][]float64 {
		out := map[string][]float64{}
		env.Reset()
		for _, bar := range ks {
			if err := env.OnBar(bar.Time, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume, bar.Quote, bar.BuyVolume, bar.TradeNum); err != nil {
				panic(err)
			}
			appendOne := func(key string, x float64) { out[key] = append(out[key], x) }
			switch name {
			case "MOM":
				appendOne(name, banta.MOM(env.Close, 2).Get(0))
			case "OBV":
				appendOne(name, banta.OBV(env.Close, env.Volume).Get(0))
			case "DEMA":
				appendOne(name, banta.DEMA(env.Close, 3).Get(0))
			case "T3":
				appendOne(name, banta.T3(env.Close, 3).Get(0))
			case "AroonOsc":
				appendOne(name, banta.AroonOsc(env.High, env.Low, 9).Get(0))
			case "StochF":
				k, d := banta.StochF(env.High, env.Low, env.Close, 5, 3)
				appendOne("STOCHF_K", k.Get(0))
				appendOne("STOCHF_D", d.Get(0))
			case "ULTOSC":
				appendOne(name, banta.ULTOSC(env.High, env.Low, env.Close, 7, 14, 28).Get(0))
			case "SAR":
				appendOne(name, banta.SAR(env.High, env.Low, .02, .2).Get(0))
			case "ROCR":
				appendOne(name, banta.ROCR(env.Close, 9).Get(0))
			case "NATR":
				appendOne(name, banta.NATR(env.High, env.Low, env.Close, 14).Get(0))
			case "TRIMA":
				appendOne(name, banta.TRIMA(env.Close, 10).Get(0))
			case "SWMA":
				appendOne(name, banta.SWMA(env.Close).Get(0))
			case "ZLMA":
				appendOne(name, banta.ZLMA(env.Close, 10).Get(0))
			case "PivotHigh":
				appendOne(name, banta.PivotHigh(env.Close, 2, 2).Get(0))
			case "PivotLow":
				appendOne(name, banta.PivotLow(env.Close, 2, 2).Get(0))
			case "AO":
				appendOne(name, banta.AO(env.High, env.Low, 5, 34).Get(0))
			case "ADOSC":
				appendOne(name, banta.ADOSC(env, 3, 10).Get(0))
			case "LINEARREG_ANGLE":
				appendOne(name, banta.LINEARREG_ANGLE(env.Close, 14).Get(0))
			case "DPO":
				appendOne(name, banta.DPO(env.Close, 20).Get(0))
			default:
				panic("unknown indicator: " + name)
			}
		}
		return out
	}

	switch in.Indicator {
	case "MOM":
		vec["MOM"] = tav.MOM(c, 2)
	case "OBV":
		vec["OBV"] = tav.OBV(c, v)
	case "DEMA":
		vec["DEMA"] = tav.DEMA(c, 3)
	case "T3":
		vec["T3"] = tav.T3(c, 3)
	case "AroonOsc":
		vec["AroonOsc"] = tav.AroonOsc(h, l, 9)
	case "StochF":
		k, d := tav.StochF(h, l, c, 5, 3)
		vec["STOCHF_K"], vec["STOCHF_D"] = k, d
	case "ULTOSC":
		vec["ULTOSC"] = tav.ULTOSC(h, l, c, 7, 14, 28)
	case "SAR":
		vec["SAR"] = tav.SAR(h, l, .02, .2)
	case "ROCR":
		vec["ROCR"] = tav.ROCR(c, 9)
	case "NATR":
		vec["NATR"] = tav.NATR(h, l, c, 14)
	case "TRIMA":
		vec["TRIMA"] = tav.TRIMA(c, 10)
	case "SWMA":
		vec["SWMA"] = tav.SWMA(c)
	case "ZLMA":
		vec["ZLMA"] = tav.ZLMA(c, 10)
	case "PivotHigh":
		vec["PivotHigh"] = tav.PivotHigh(c, 2, 2)
	case "PivotLow":
		vec["PivotLow"] = tav.PivotLow(c, 2, 2)
	case "AO":
		vec["AO"] = tav.AO(h, l, 5, 34)
	case "ADOSC":
		vec["ADOSC"] = tav.ADOSC(h, l, c, v, 3, 10)
	case "LINEARREG_ANGLE":
		vec["LINEARREG_ANGLE"] = tav.LINEARREG_ANGLE(c, 14)
	case "DPO":
		vec["DPO"] = tav.DPO(c, 20)
	default:
		panic("unknown indicator: " + in.Indicator)
	}
	env, err := banta.NewBarEnv("compare", "spot", "", "1d")
	if err != nil {
		panic(err)
	}
	state := stateRun(env, in.Indicator)
	cleanVec, cleanState := map[string][]interface{}{}, map[string][]interface{}{}
	for k, x := range vec {
		cleanVec[k] = clean(x)
	}
	for k, x := range state {
		cleanState[k] = clean(x)
	}
	result, err := json.Marshal(output{Indicator: in.Indicator, Vector: cleanVec, State: cleanState})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(result))
}

func extract(ks []banta.Kline) (open, high, low, close, volume []float64, times []int64) {
	open, high, low, close, volume, times = make([]float64, len(ks)), make([]float64, len(ks)), make([]float64, len(ks)), make([]float64, len(ks)), make([]float64, len(ks)), make([]int64, len(ks))
	for i, k := range ks {
		open[i], high[i], low[i], close[i], volume[i], times[i] = k.Open, k.High, k.Low, k.Close, k.Volume, k.Time
	}
	return
}

func canonical(name string) string {
	switch strings.ToLower(name) {
	case "ao":
		return "AO"
	case "adosc":
		return "ADOSC"
	case "linearreg_angle", "linearregangle":
		return "LINEARREG_ANGLE"
	case "dpo":
		return "DPO"
	case "mom":
		return "MOM"
	case "obv":
		return "OBV"
	case "dema":
		return "DEMA"
	case "t3":
		return "T3"
	case "aroonosc":
		return "AroonOsc"
	case "stochf", "stochf_k", "stochf_d":
		return "StochF"
	case "ultosc":
		return "ULTOSC"
	case "sar", "psar":
		return "SAR"
	case "rocr":
		return "ROCR"
	case "natr":
		return "NATR"
	case "trima":
		return "TRIMA"
	case "swma":
		return "SWMA"
	case "zlma", "zema":
		return "ZLMA"
	case "pivothigh":
		return "PivotHigh"
	case "pivotlow":
		return "PivotLow"
	default:
		return name
	}
}
