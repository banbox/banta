package tav

func HL2(a, b []float64) []float64 {
	res := make([]float64, len(a))
	for i, va := range a {
		res[i] = va*0.5 + b[i]*0.5
	}
	return res
}

func HLC3(a, b, c []float64) []float64 {
	res := make([]float64, len(a))
	for i, va := range a {
		res[i] = (va + b[i] + c[i]) / 3
	}
	return res
}
