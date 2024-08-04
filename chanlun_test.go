package banta

import (
	"fmt"
	"strings"
	"testing"
)

type SegPair struct {
	Name string
	Pens []float64    // 代表所有笔的端点，任意相邻两个数构成一笔
	Segs [][2]float64 // 所有线段端点，代表有n-1个线段
}

func TestBuildSegs(t *testing.T) {
	setDebug(false)
	items := []*SegPair{
		//// 下面来自图解缠论1.7的图1-22和1-23
		//{
		//	Name: "1-22",
		//	Pens: []float64{10, 30, 20, 50, 40, 60, 55, 150, 110, 130, 105, 120, 108, 118, 80, 127, 105, 130, 104, 160, 155, 180, 150},
		//	Segs: [][2]float64{{1, 10}, {8, 150}, {15, 80}, {22, 180}},
		//},
		//{
		//	Name: "1-23",
		//	Pens: []float64{100, 95, 110, 90, 112, 60, 70, 63, 80, 63, 72, 20, 30, 10, 27, 11, 25, 5},
		//	Segs: [][2]float64{{1, 100}, {18, 5}},
		//},
		//{
		//	Name: "1-24",
		//	Pens: []float64{100, 50, 70, 45, 60, 25, 55, 24, 40, 24, 45, 10, 23, 15, 31, 25, 63, 43, 53, 40, 75, 45, 105, 85, 100, 60, 70, 67, 98, 75},
		//	Segs: [][2]float64{{1, 100}, {12, 10}, {23, 105}, {26, 60}, {29, 98}},
		//},
		//{
		//	Name: "1-25",
		//	Pens: []float64{10, 50, 40, 46, 39, 45, 38, 55, 43, 65, 46, 70, 42, 55, 43, 100, 90}, // 需要额外一个笔确认最后
		//	Segs: [][2]float64{{1, 10}, {16, 100}},
		//},
		//{
		//	Name: "1-25-2",
		//	Pens: []float64{10, 50, 40, 46, 39, 45, 38, 55, 43, 65, 46, 70, 42, 55, 41, 100},
		//	Segs: [][2]float64{{1, 10}, {12, 70}, {15, 41}},
		//},
		//// 下面来自图解缠论1.7的图1-28
		//{ // a
		//	Name: "a",
		//	Pens: []float64{10, 30, 20, 40},
		//	Segs: [][2]float64{{1, 10}, {4, 40}},
		//},
		//{ // b
		//	Name: "b",
		//	Pens: []float64{10, 30, 20, 40, 25, 50},
		//	Segs: [][2]float64{{1, 10}, {6, 50}},
		//},
		//{ // c
		//	Name: "c",
		//	Pens: []float64{10, 30, 25, 60, 50, 70, 40, 90, 80, 120},
		//	Segs: [][2]float64{{1, 10}, {10, 120}},
		//},
		//{ // d
		//	Name: "d",
		//	Pens: []float64{10, 30, 25, 50, 35, 45, 38, 60, 40, 80, 70, 90},
		//	Segs: [][2]float64{{1, 10}, {12, 90}},
		//},
		//{ // e 5
		//	Name: "e",
		//	Pens: []float64{10, 30, 20, 50, 25, 70, 60, 90},
		//	Segs: [][2]float64{{1, 10}, {8, 90}},
		//},
		//{ // f
		//	Name: "f",
		//	Pens: []float64{10, 30, 20, 60, 40, 60, 45, 60, 50, 80},
		//	Segs: [][2]float64{{1, 10}, {10, 80}},
		//},
		//{ // g
		//	Name: "g",
		//	Pens: []float64{10, 30, 20, 29, 22, 28, 23, 40, 21, 60, 50, 70},
		//	Segs: [][2]float64{{1, 10}, {12, 70}},
		//},
		//{ // h
		//	Name: "h",
		//	Pens: []float64{10, 30, 25, 31, 26, 40, 27, 60, 50, 59, 51, 58, 53, 70, 65}, // 最后需要额外一个笔确认
		//	Segs: [][2]float64{{1, 10}, {14, 70}},
		//},
		//{ // i
		//	Name: "i",
		//	Pens: []float64{10, 30, 20, 50, 25, 70, 60, 80},
		//	Segs: [][2]float64{{1, 10}, {8, 80}},
		//},
		//{ // j 10
		//	Name: "j",
		//	Pens: []float64{10, 30, 20, 28, 18, 50, 35, 70},
		//	Segs: [][2]float64{{1, 10}, {8, 70}},
		//},
		//{ // k
		//	Name: "k",
		//	Pens: []float64{10, 30, 20, 50, 40, 48, 35, 60, 49, 70, 65},
		//	Segs: [][2]float64{{1, 10}, {10, 70}},
		//},
		//{ // l
		//	Name: "l",
		//	Pens: []float64{10, 30, 20, 50, 28, 70, 60, 90, 80, 110, 85, 130, 120, 140},
		//	Segs: [][2]float64{{1, 10}, {14, 140}},
		//},
		//// 下面三个来自图解缠论1.7的图1-29
		//{
		//	Name: "1-29-1",
		//	Pens: []float64{10, 30, 20, 70, 55, 65, 50, 60, 56, 90, 80},
		//	Segs: [][2]float64{{1, 10}, {4, 70}, {7, 50}, {10, 90}},
		//},
		//{
		//	Name: "1-29-2",
		//	Pens: []float64{10, 30, 20, 70, 55, 65, 50, 65, 56, 90, 80},
		//	Segs: [][2]float64{{1, 10}, {10, 90}},
		//},
		//{
		//	Name: "1-29-3",
		//	Pens: []float64{10, 30, 20, 70, 55, 65, 50, 67, 56, 90, 80},
		//	Segs: [][2]float64{{1, 10}, {10, 90}},
		//},
		//// 下面三个来自图解缠论1.7的图1-30
		//{
		//	Name: "1-30-1",
		//	Pens: []float64{100, 80, 90, 70, 85, 60, 78, 50, 130, 80, 88, 58, 72, 40, 55, 30, 47, 36, 60},
		//	Segs: [][2]float64{{1, 100}, {16, 30}, {19, 60}},
		//},
		//{
		//	Name: "1-30-2",
		//	Pens: []float64{100, 80, 90, 70, 85, 60, 78, 50, 130, 80, 88, 58, 84, 70, 89, 86}, // 最后额外需要一笔验证
		//	Segs: [][2]float64{{1, 100}, {8, 50}, {15, 89}},
		//},
		//{
		//	// 这里和书中不一致，书中在12出划分线段，需要往后看9根K线，太多了；这里简单起见保持目前逻辑
		//	Name: "1-31",
		//	Pens: []float64{100, 90, 95, 80, 85, 60, 75, 65, 70, 60, 67, 40, 90, 64, 70, 50, 60, 41, 55, 52, 65, 38, 52, 10},
		//	Segs: [][2]float64{{1, 100}, {18, 41}, {21, 65}, {24, 10}},
		//},
		//{
		//	Name: "1-32-1",
		//	Pens: []float64{50, 40, 48, 35, 70, 40, 50, 45, 65, 35, 43, 10},
		//	Segs: [][2]float64{{1, 50}, {4, 35}, {9, 65}, {12, 10}},
		//},
		//{
		//	// 这里和书中不一致，但感觉目前的输出更合理
		//	Name: "1-32-2",
		//	Pens: []float64{70, 20, 40, 15, 65, 20, 50, 35, 60, 38, 55, 10},
		//	Segs: [][2]float64{{1, 70}, {4, 15}, {9, 60}, {12, 10}},
		//},
		//// 下面来自教你炒股票之缠论新解第五章
		//{
		//	Name: "5-3-1",
		//	Pens: []float64{10, 30, 15, 60, 50, 80, 40, 70, 55, 65, 35},
		//	Segs: [][2]float64{{1, 10}, {6, 80}, {11, 35}},
		//},
		//{
		//	Name: "5-3-2",
		//	Pens: []float64{10, 30, 15, 60, 50, 80, 53, 65, 55, 64, 56, 90, 85},
		//	Segs: [][2]float64{{1, 10}, {12, 90}},
		//},
		//{
		//	Name: "5-3-3",
		//	Pens: []float64{10, 30, 20, 60, 50, 90, 75, 80, 55, 120, 85, 95, 49},
		//	Segs: [][2]float64{{1, 10}, {10, 120}, {13, 49}},
		//},
		//{
		//	Name: "5-3-4", // 书中分成3段错误，因第一段有缺口，第二段必须底分型确认
		//	Pens: []float64{10, 30, 20, 60, 50, 90, 80, 85, 55, 87, 83, 120, 60},
		//	Segs: [][2]float64{{1, 10}, {12, 120}},
		//},
		// 自定义数据
		{
			Name: "eth-30m-20240619",
			Pens: []float64{3670, 3691, 3684, 3715, 3684, 3711, 3695, 3710, 3642, 3715, 3651, 3679, 3501,
				3543, 3523, 3547, 3427, 3498, 3489, 3510, 3462, 3657, 3512, 3574, 3476, 3515, 3472, 3541,
				3426, 3492, 3461, 3529, 3361, 3529, 3472, 3557},
			Segs: [][2]float64{{1, 3670}},
		},
		//{
		//	Name: "eth-15m-240613",
		//	Pens: []float64{3426.01, 3487.96, 3461.03, 3486.68, 3466.06, 3528.93, 3514.22, 3527.1, 3507.55,
		//		3525.2, 3361.05, 3529, 3426.23, 3502.5, 3472.34, 3557.54, 3526.05, 3540.25, 3526.7, 3594.99,
		//		3556.8, 3574.98, 3534.58, 3574.88, 3553.2, 3572.95, 3540},
		//	Segs: [][2]float64{{1, 3426.01}, {6, 3528.93}, {11, 3361.05}, {20, 3594.99}, {23, 3534.58}},
		//},
	}
	for _, it := range items {
		cg := &CGraph{}
		dirt := float64(-1)
		if it.Pens[1] < it.Pens[0] {
			// 顶分型
			dirt = 1
		}
		var pv *CPoint
		for i, p := range it.Pens {
			if i > 0 && dirt*(p-it.Pens[i-1]) <= 0 {
				panic(fmt.Sprintf("sample[%v] idx: %v invalid", it.Name, i))
			}
			pt := cg.NewPoint(dirt, p, i+1)
			if pv != nil {
				pen := pv.PenTo(pt)
				pen.State = CLDone
				cg.AddPen(pen)
			}
			pv = pt
			dirt *= -1
		}
		// 检查线段是否一致
		match := -1
		for i, s := range cg.Segs {
			if i >= len(it.Segs)-1 {
				match = i
				break
			}
			req := it.Segs[i+1]
			if s.End.BarId != int(req[0]) || s.End.Price != req[1] {
				match = i
				break
			} else if i == 0 {
				req = it.Segs[0]
				if s.Start.BarId != int(req[0]) || s.Start.Price != req[1] {
					match = i
					break
				}
			}
		}
		if match >= 0 {
			texts := make([]string, 0, len(cg.Segs))
			for i, s := range cg.Segs {
				if i == 0 {
					texts = append(texts, s.Start.StrPoint())
				}
				texts = append(texts, s.End.StrPoint())
			}
			t.Errorf("sample[%v] fail %v, gen: %v", it.Name, match, strings.Join(texts, " "))
		}
	}
}