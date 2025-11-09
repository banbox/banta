package banta

import (
	"errors"
	"fmt"
	"math"
	"strconv"
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

func RemoveFromArr[T comparable](arr []T, it T, num int) []T {
	res := make([]T, 0, len(arr))
	for _, v := range arr {
		if v == it && (num < 0 || num > 0) {
			num -= 1
			continue
		}
		res = append(res, v)
	}
	return res
}

const (
	SecsMin  = 60
	SecsHour = SecsMin * 60
	SecsDay  = SecsHour * 24
	SecsWeek = SecsDay * 7
	SecsMon  = SecsDay * 30
	SecsQtr  = SecsMon * 3
	SecsYear = SecsDay * 365
)

var (
	errTfTooShort = errors.New("timeframe string too short")
)

func ParseTimeFrame(timeframe string) (int, error) {
	if len(timeframe) < 2 {
		return 0, errTfTooShort
	}

	amountStr := timeframe[:len(timeframe)-1]
	unit := timeframe[len(timeframe)-1]

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return 0, err
	}

	var scale int
	switch unit {
	case 'y', 'Y':
		scale = SecsYear
	case 'q', 'Q':
		scale = SecsQtr
	case 'M':
		scale = SecsMon
	case 'w', 'W':
		scale = SecsWeek
	case 'd', 'D':
		scale = SecsDay
	case 'h', 'H':
		scale = SecsHour
	case 'm':
		scale = SecsMin
	case 's', 'S':
		scale = 1
	default:
		return 0, fmt.Errorf("timeframe unit %v is not supported", string(unit))
	}

	return amount * scale, nil
}
