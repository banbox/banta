package banta

import (
	"slices"
	"strconv"
	"strings"
	"testing"
)

type CaseItem struct {
	Title   string
	Expects []float64
	RunVec  func(o, h, l, c, v, i []float64) []float64
	Run     func(e *BarEnv) float64
}

var env = &BarEnv{
	TimeFrame:  "1d",
	TFMSecs:    86400000,
	Exchange:   "binance",
	MarketType: "future",
}

func runIndCases(t *testing.T, items []CaseItem) {
	var fails = make(map[string]int)
	var results = make(map[string][]float64)
	fakeBarNum := len(DataKline)
	for _, it := range items {
		fails[it.Title] = 0
		results[it.Title] = make([]float64, fakeBarNum)
	}
	RunFakeEnv(env, DataKline, func(i int, kline Kline) {
		for _, it := range items {
			calcVal := it.Run(env)
			results[it.Title][i] = calcVal
			if !equalNearly(calcVal, it.Expects[i]) {
				fails[it.Title] += 1
			}
		}
	})
	for _, it := range items {
		failNum, _ := fails[it.Title]
		calcus, _ := results[it.Title]
		if failNum == 0 {
			t.Logf("pass %s: %v", it.Title, arrToStr(calcus))
		} else {
			t.Errorf("FAIL %d %s\nExpect: %v\nCalcul: %v", failNum, it.Title, arrToStr(it.Expects), arrToStr(calcus))
		}
	}
	t.Log("start test vector indicators")
	o, h, l, c, v, i := extractOHLCV(DataKline)
	for _, it := range items {
		calcus := it.RunVec(o, h, l, c, v, i)
		failNum := 0
		if len(it.Expects) != len(calcus) {
			failNum = len(calcus)
		} else {
			for j, expVal := range it.Expects {
				if !equalNearly(calcus[j], expVal) {
					failNum += 1
				}
			}
		}
		if failNum == 0 {
			t.Logf("pass %s: %v", it.Title, arrToStr(calcus))
		} else {
			t.Errorf("FAIL %d %s\nExpect: %v\nCalcul: %v", failNum, it.Title, arrToStr(it.Expects), arrToStr(calcus))
		}
	}
}

func extractOHLCV(klineData []Kline) (o, h, l, c, v, i []float64) {
	barNum := len(klineData)
	o = make([]float64, 0, barNum)
	h = make([]float64, 0, barNum)
	l = make([]float64, 0, barNum)
	c = make([]float64, 0, barNum)
	v = make([]float64, 0, barNum)
	i = make([]float64, 0, barNum)
	for _, k := range klineData {
		o = append(o, k.Open)
		h = append(h, k.High)
		l = append(l, k.Low)
		c = append(c, k.Close)
		v = append(v, k.Volume)
		i = append(i, k.Info)
	}
	return
}

func arrToStr(arr []float64) string {
	var b strings.Builder
	b.Grow(len(arr) * 4)
	b.WriteString("[")
	for _, v := range arr {
		b.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
		b.WriteString(", ")
	}
	b.WriteString("]")
	return b.String()
}

func TestSeries(t *testing.T) {
	closeArr := []float64{30573.6, 30612.7, 31149, 30756.1, 30488.4, 29874.4, 30327.9, 30269.3, 30147.8, 30396.9, 30608.4, 30368.9, 31441.7, 30293.3, 30276.4, 30216.8, 30126.1, 29845.6, 29895.5, 29791, 29891.4, 29783.5, 30070.8, 29163.8, 29216.3, 29336, 29209.7, 29299.9, 29339.1, 29271.1, 29220.8, 29701.2, 29170.1, 29180.2, 29101.1, 29057.7, 29075.9, 29202.7, 29759, 29572.8, 29443.7, 29415.5, 29420.7, 29293.3, 29419.5, 29188.8, 28714.4, 26609.7, 26042.1, 26088.3, 26175.9, 26115.4, 26044.4, 26419.2, 26164.6, 26051.7, 26004.3, 26087.7}
	closeAdd12 := []float64{30573.719999999998, 30612.82, 31149.12, 30756.219999999998, 30488.52, 29874.52, 30328.02, 30269.42, 30147.92, 30397.02, 30608.52, 30369.02, 31441.82, 30293.42, 30276.52, 30216.92, 30126.219999999998, 29845.719999999998, 29895.62, 29791.12, 29891.52, 29783.62, 30070.92, 29163.92, 29216.42, 29336.12, 29209.82, 29300.02, 29339.219999999998, 29271.219999999998, 29220.92, 29701.32, 29170.219999999998, 29180.32, 29101.219999999998, 29057.82, 29076.02, 29202.82, 29759.12, 29572.92, 29443.82, 29415.62, 29420.82, 29293.42, 29419.62, 29188.92, 28714.52, 26609.82, 26042.219999999998, 26088.42, 26176.02, 26115.52, 26044.52, 26419.32, 26164.719999999998, 26051.82, 26004.42, 26087.82}
	closeSub_1 := []float64{30573.5, 30612.600000000002, 31148.9, 30756, 30488.300000000003, 29874.300000000003, 30327.800000000003, 30269.2, 30147.7, 30396.800000000003, 30608.300000000003, 30368.800000000003, 31441.600000000002, 30293.2, 30276.300000000003, 30216.7, 30126, 29845.5, 29895.4, 29790.9, 29891.300000000003, 29783.4, 30070.7, 29163.7, 29216.2, 29335.9, 29209.600000000002, 29299.800000000003, 29339, 29271, 29220.7, 29701.100000000002, 29170, 29180.100000000002, 29101, 29057.600000000002, 29075.800000000003, 29202.600000000002, 29758.9, 29572.7, 29443.600000000002, 29415.4, 29420.600000000002, 29293.2, 29419.4, 29188.7, 28714.300000000003, 26609.600000000002, 26042, 26088.2, 26175.800000000003, 26115.300000000003, 26044.300000000003, 26419.100000000002, 26164.5, 26051.600000000002, 26004.2, 26087.600000000002}
	closeMul1_1 := []float64{33630.96, 33673.97, 34263.9, 33831.71, 33537.240000000005, 32861.840000000004, 33360.69, 33296.23, 33162.58, 33436.590000000004, 33669.240000000005, 33405.79, 34585.87, 33322.630000000005, 33304.04, 33238.48, 33138.71, 32830.16, 32885.05, 32770.100000000006, 32880.54, 32761.850000000002, 33077.880000000005, 32080.18, 32137.93, 32269.600000000002, 32130.670000000002, 32229.890000000003, 32273.010000000002, 32198.210000000003, 32142.88, 32671.320000000003, 32087.11, 32098.220000000005, 32011.210000000003, 31963.470000000005, 31983.490000000005, 32122.970000000005, 32734.9, 32530.08, 32388.070000000003, 32357.050000000003, 32362.770000000004, 32222.63, 32361.450000000004, 32107.68, 31585.840000000004, 29270.670000000002, 28646.31, 28697.13, 28793.490000000005, 28726.940000000002, 28648.840000000004, 29061.120000000003, 28781.06, 28656.870000000003, 28604.730000000003, 28696.470000000005}
	absLSubH := []float64{356.90000000000146, 650.0999999999985, 835.7999999999993, 719.4000000000015, 699.4000000000015, 1750, 763.5999999999985, 353.2000000000007, 401.40000000000146, 1111.2000000000007, 543.5, 794.5, 1617, 1763.4000000000015, 159.79999999999927, 391.59999999999854, 699.7000000000007, 827.5, 433.5, 951.2000000000007, 350.2999999999993, 383.7999999999993, 649.2999999999993, 1261.7000000000007, 346.5, 587.5, 492.90000000000146, 422.90000000000146, 160.40000000000146, 436.5, 425.7999999999993, 1185, 1153.6000000000022, 484.59999999999854, 543.7999999999993, 185, 220.5, 592.2000000000007, 1117.5999999999985, 787.7000000000007, 426.09999999999854, 345.09999999999854, 110, 206.40000000000146, 616.7000000000007, 442.09999999999854, 552.3000000000029, 4194.9000000000015, 1218, 486, 336.5, 480, 848.2000000000007, 1006, 733.2999999999993, 545.5999999999985, 160.40000000000146, 218}
	crossArr := []float64{0, 0, 0, 0, 0, -1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, -1, -2, -3, -4, -5, 1, -1, -2, -3, -4, -5, -6, -7, -8, -9, -10, -11, -12, -13, -14, -15, -16, -17, -18, -19, -20, -21, -22, -23, -24, -25, -26, -27, -28, -29, -30, -31, -32, -33, -34, -35}
	rangeArr := []float64{30573.6, 30612.7, 31149, 31149, 31149, 30756.1, 30488.4, 30327.9, 30327.9, 30396.9, 30608.4, 30608.4, 31441.7, 31441.7, 31441.7, 30293.3, 30276.4, 30216.8, 30126.1, 29895.5, 29895.5, 29891.4, 30070.8, 30070.8, 30070.8, 29336, 29336, 29336, 29339.1, 29339.1, 29339.1, 29701.2, 29701.2, 29701.2, 29180.2, 29180.2, 29101.1, 29202.7, 29759, 29759, 29759, 29572.8, 29443.7, 29420.7, 29420.7, 29419.5, 29419.5, 29188.8, 28714.4, 26609.7, 26175.9, 26175.9, 26175.9, 26419.2, 26419.2, 26419.2, 26164.6, 26087.7}
	runIndCases(t, []CaseItem{
		{"close", closeArr, nil, func(env *BarEnv) float64 {
			return env.Close.Get(0)
		}},
		{"close+0.12", closeAdd12, nil, func(env *BarEnv) float64 {
			return env.Close.Add(0.12).Get(0)
		}},
		{"close-0.1", closeSub_1, nil, func(env *BarEnv) float64 {
			return env.Close.Sub(0.1).Get(0)
		}},
		{"close*1.1", closeMul1_1, nil, func(env *BarEnv) float64 {
			return env.Close.Mul(1.1).Get(0)
		}},
		{"abs(l-h)", absLSubH, nil, func(env *BarEnv) float64 {
			return env.Low.Sub(env.High).Abs().Get(0)
		}},
		{"cross", crossArr, nil, func(env *BarEnv) float64 {
			return float64(Cross(env.Close, 30000))
		}},
		{"range", rangeArr, nil, func(env *BarEnv) float64 {
			arr := env.Close.Range(0, 3)
			return slices.Max(arr)
		}},
	})
}

// 新的测试运行器：对比带状态版本和并行版本的结果
func runAndCompareCases(t *testing.T, klineData []Kline, items []CaseItem, showTrue bool) {
	o, h, l, c, v, iData := extractOHLCV(klineData)

	for _, it := range items {
		t.Run(it.Title, func(t *testing.T) {
			// 运行并行版本
			vecResults := it.RunVec(o, h, l, c, v, iData)

			// 运行带状态版本
			localEnv := &BarEnv{} // 使用本地env确保测试隔离
			*localEnv = *env      // 复制全局env的配置
			var stateResults []float64

			// 模拟K线推送
			RunFakeEnv(localEnv, klineData, func(i int, kline Kline) {
				raw := it.Run(localEnv)
				stateResults = append(stateResults, raw)
			})

			if len(vecResults) != len(stateResults) {
				t.Errorf("FAIL: Result count mismatch. Vector: %d, State: %d", len(vecResults), len(stateResults))
				return
			}

			// 对比结果
			isMismatch := false
			for i := range vecResults {
				if !equalNearly(vecResults[i], stateResults[i]) {
					isMismatch = true
					break
				}
			}

			if isMismatch {
				t.Errorf("State  Result: %s", arrToStr(stateResults))
				t.Errorf("Vector Result: %s", arrToStr(vecResults))
			} else if showTrue {
				t.Logf("%s", arrToStr(vecResults))
			}
		})
	}
}
