package banta

import (
	"fmt"
	"math"
)

const thresFloat64Eq = 1e-9

/*
numSign 获取数字的方向；1，-1或0
*/
func numSign(obj interface{}) int {
	if val, ok := obj.(int); ok {
		if val > 0 {
			return 1
		} else if val < 0 {
			return -1
		} else {
			return 0
		}
	} else if val, ok := obj.(float32); ok {
		if val > 0 {
			return 1
		} else if val < 0 {
			return -1
		} else {
			return 0
		}
	} else if val, ok := obj.(float64); ok {
		if val > 0 {
			return 1
		} else if val < 0 {
			return -1
		} else {
			return 0
		}
	} else {
		panic(fmt.Errorf("invalid type for numSign: %t", obj))
	}
}

/*
equalNearly 判断两个float是否近似相等，解决浮点精读导致不等
*/
func equalNearly(a, b float64) bool {
	return equalIn(a, b, thresFloat64Eq)
}

/*
equalIn 判断两个float是否在一定范围内近似相等
*/
func equalIn(a, b, thres float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return math.Abs(a-b) <= thres
}
