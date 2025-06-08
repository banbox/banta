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
	for i := 0; i < b.N; i++ {
		data := generateRandomSlice(1000)
		Aroon(data, data, 10)
	}
}

func BenchmarkAroon2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		data := generateRandomSlice(1000)
		Aroon(data, data, 10)
	}
}

// 生成随机测试数据
func generateTestData(size int) []float64 {
	data := make([]float64, size)
	for i := range data {
		data[i] = rand.Float64() * 100 // 生成0-100之间的随机数
	}
	return data
}
