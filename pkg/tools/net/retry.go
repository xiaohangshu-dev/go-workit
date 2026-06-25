package net

import (
	"math"
	"math/rand"
	"time"
)

// RetryOptions 重试策略配置
type RetryOptions struct {
	// MaxRetries 最大重试次数，0 表示不重试
	MaxRetries int
	// BaseDelay 初始延迟
	BaseDelay time.Duration
	// MaxDelay 最大延迟
	MaxDelay time.Duration
	// RetryOn 触发重试的 HTTP 状态码，为空时重试所有 4xx/5xx
	RetryOn []int
}

func (o *RetryOptions) normalize() {
	if o.BaseDelay <= 0 {
		o.BaseDelay = 100 * time.Millisecond
	}
	if o.MaxDelay <= 0 {
		o.MaxDelay = 2 * time.Second
	}
	if len(o.RetryOn) == 0 {
		o.RetryOn = []int{408, 429, 500, 502, 503}
	}
}

// shouldRetry 判断指定状态码是否需要重试
func (o *RetryOptions) shouldRetry(statusCode int) bool {
	for _, code := range o.RetryOn {
		if statusCode == code {
			return true
		}
	}
	return false
}

// retryDelay 计算第 N 次重试的延迟时间（指数退避 + 随机抖动）
func retryDelay(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	if attempt <= 0 {
		return 0
	}
	delay := float64(baseDelay) * math.Pow(2, float64(attempt-1))
	delay = math.Min(delay, float64(maxDelay))
	// 添加 0-25% 随机抖动
	jitter := rand.Float64() * delay * 0.25
	return time.Duration(delay + jitter)
}
