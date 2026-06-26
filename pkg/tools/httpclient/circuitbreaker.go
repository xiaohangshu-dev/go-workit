package httpclient

import (
	"sync"
	"sync/atomic"
	"time"
)

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	CircuitBreakerClosed   CircuitBreakerState = iota // 关闭（正常）
	CircuitBreakerOpen                                // 断开（拒绝请求）
	CircuitBreakerHalfOpen                            // 半开（试探恢复）
)

// CircuitBreakerOptions 熔断器配置
type CircuitBreakerOptions struct {
	// FailureThreshold 连续失败次数阈值，达到后熔断器断开
	FailureThreshold int
	// SuccessThreshold 半开后连续成功次数阈值，达到后熔断器关闭
	SuccessThreshold int
	// HalfOpenMaxRequests 半开状态最大请求数
	HalfOpenMaxRequests int
	// OpenDuration 断开持续时间，之后进入半开状态
	OpenDuration time.Duration
}

func (o *CircuitBreakerOptions) normalize() {
	if o.FailureThreshold <= 0 {
		o.FailureThreshold = 5
	}
	if o.SuccessThreshold <= 0 {
		o.SuccessThreshold = 2
	}
	if o.HalfOpenMaxRequests <= 0 {
		o.HalfOpenMaxRequests = 1
	}
	if o.OpenDuration <= 0 {
		o.OpenDuration = 30 * time.Second
	}
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	options      CircuitBreakerOptions
	state        atomic.Int32
	failureCount atomic.Int32
	successCount atomic.Int32
	openTime     atomic.Int64 // 断开时的时间戳
	halfOpenPerm atomic.Int32 // 半开状态已分配的许可数
	mu           sync.Mutex
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(options CircuitBreakerOptions) *CircuitBreaker {
	options.normalize()
	return &CircuitBreaker{options: options}
}

// State 获取当前状态
func (cb *CircuitBreaker) State() CircuitBreakerState {
	return CircuitBreakerState(cb.state.Load())
}

// AllowRequest 判断是否允许请求通过
func (cb *CircuitBreaker) AllowRequest() bool {
	state := cb.State()

	switch state {
	case CircuitBreakerClosed:
		return true

	case CircuitBreakerOpen:
		// 检查是否需要切换到半开
		if time.Since(time.Unix(0, cb.openTime.Load())) > cb.options.OpenDuration {
			if cb.state.CompareAndSwap(int32(CircuitBreakerOpen), int32(CircuitBreakerHalfOpen)) {
				cb.successCount.Store(0)
				cb.halfOpenPerm.Store(0)
				return true
			}
		}
		return false

	case CircuitBreakerHalfOpen:
		// 限制半开状态请求数
		perm := cb.halfOpenPerm.Add(1)
		if perm <= int32(cb.options.HalfOpenMaxRequests) {
			return true
		}
		return false

	default:
		return true
	}
}

// OnSuccess 记录成功
func (cb *CircuitBreaker) OnSuccess() {
	switch cb.State() {
	case CircuitBreakerHalfOpen:
		count := cb.successCount.Add(1)
		if count >= int32(cb.options.SuccessThreshold) {
			cb.state.Store(int32(CircuitBreakerClosed))
			cb.failureCount.Store(0)
			cb.successCount.Store(0)
		}
	case CircuitBreakerClosed:
		cb.failureCount.Store(0)
	}
}

// OnFailure 记录失败
func (cb *CircuitBreaker) OnFailure() {
	switch cb.State() {
	case CircuitBreakerHalfOpen:
		// 半开状态失败一次立即回到断开
		cb.state.Store(int32(CircuitBreakerOpen))
		cb.openTime.Store(time.Now().UnixNano())
		cb.successCount.Store(0)
	case CircuitBreakerClosed:
		count := cb.failureCount.Add(1)
		if count >= int32(cb.options.FailureThreshold) {
			cb.state.Store(int32(CircuitBreakerOpen))
			cb.openTime.Store(time.Now().UnixNano())
		}
	}
}
