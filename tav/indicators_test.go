package tav

import (
	"math/rand"
	"testing"
)

// 生成随机数组
func generateRandomSlice(n int) []float64 {
	slice := make([]float64, n)
	r := rand.New(rand.NewSource(42))
	for i := range slice {
		slice[i] = r.Float64()
	}
	return slice
}

func BenchmarkAroon(b *testing.B) {
	data := generateRandomSlice(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Aroon(data, data, 10)
	}
}
