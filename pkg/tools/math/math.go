package math

import (
	"math"
	"math/rand"
	"time"
)

// 返回两个整数中的最大值
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 返回两个整数中的最小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 返回整数的绝对值
func Abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// 限制数值范围
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// 四舍五入
func Round(value float64) float64 {
	return math.Round(value)
}

// 计算平方根
func Sqrt(value float64) float64 {
	return math.Sqrt(value)
}

// 计算幂
func Pow(base, exp float64) float64 {
	return math.Pow(base, exp)
}

// 生成随机整数
func RandomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

// init 初始化随机数种子
func init() {
	rand.Seed(time.Now().UnixNano())
}
