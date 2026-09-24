package banta

import (
	"math"
	"testing"
)

var cacheKeyBenchSink int

func oldMACDKey(fast, slow, smooth, initType int) int {
	return fast*1000 + slow*100 + smooth*10 + initType
}

func currentMACDKey(fast, slow, smooth, initType int) int {
	return fast*100000000 + slow*1000000 + smooth*100 + initType
}

func oldFloatKey(period int, stdUp, stdDn float64) int {
	return period*10000 + int(stdUp*1000) + int(stdDn*10)
}

func BenchmarkCacheKeyCurrentInts(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cacheKeyBenchSink = currentMACDKey(12, 26, 9, 0)
	}
}

func BenchmarkCacheKeyOldInts(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cacheKeyBenchSink = oldMACDKey(12, 26, 9, 0)
	}
}

func BenchmarkCacheKeyCurrentFloats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cacheKeyBenchSink = pkey(ikey(20), fkey(2.0), fkey(2.0))
	}
}

func BenchmarkCacheKeyOldFloats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cacheKeyBenchSink = oldFloatKey(20, 2.0, 2.0)
	}
}

func BenchmarkCacheKeyCurrentSTC(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cacheKeyBenchSink = pkey(ikey(12), ikey(26), ikey(50), fkey(0.5))
	}
}

func BenchmarkCacheKeyCurrentRMA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cacheKeyBenchSink = pkey(ikey(20), ikey(0), fkey(math.NaN()))
	}
}
