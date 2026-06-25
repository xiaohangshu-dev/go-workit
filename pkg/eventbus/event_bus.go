// Package eventbus 提供通用的事件总线实现，支持事件的发布与订阅。
//
// 使用方式：
//
//	bus := eventbus.New()
//	bus.Subscribe("order.created", handler)
//	bus.Publish(ctx, "order.created", event)
package eventbus

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// Event 定义事件接口
type Event interface {
	// EventName 返回事件名称
	EventName() string
}

// Handler 定义事件处理器
type Handler interface {
	// Handle 处理事件
	Handle(ctx context.Context, event Event) error
}

// HandlerFunc 是一个函数类型，实现了 Handler 接口
type HandlerFunc func(ctx context.Context, event Event) error

// Handle 实现 Handler 接口
func (f HandlerFunc) Handle(ctx context.Context, event Event) error {
	return f(ctx, event)
}

// subscription 表示一个订阅，包含处理器和唯一标识
type subscription struct {
	id      uint64
	handler Handler
}

// Bus 事件总线
type Bus struct {
	mu      sync.RWMutex
	subs    map[string][]*subscription
	counter atomic.Uint64
}

// New 创建一个新的事件总线
func New() *Bus {
	return &Bus{
		subs: make(map[string][]*subscription),
	}
}

// Subscribe 订阅指定事件，返回订阅ID可用于取消订阅
func (b *Bus) Subscribe(eventName string, handler Handler) uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.counter.Add(1)
	b.subs[eventName] = append(b.subs[eventName], &subscription{
		id:      id,
		handler: handler,
	})
	return id
}

// SubscribeFunc 使用函数订阅指定事件，返回订阅ID可用于取消订阅
func (b *Bus) SubscribeFunc(eventName string, fn func(ctx context.Context, event Event) error) uint64 {
	return b.Subscribe(eventName, HandlerFunc(fn))
}

// Publish 发布事件，同步调用所有已注册的处理器
func (b *Bus) Publish(ctx context.Context, event Event) error {
	eventName := event.EventName()

	b.mu.RLock()
	subs, exists := b.subs[eventName]
	b.mu.RUnlock()

	if !exists {
		return nil
	}

	var errs []error
	for _, sub := range subs {
		if err := sub.handler.Handle(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("eventbus: %d handler(s) failed for event %q: %w", len(errs), eventName, errs[0])
	}
	return nil
}

// Unsubscribe 根据订阅ID取消订阅
func (b *Bus) Unsubscribe(eventName string, id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs, exists := b.subs[eventName]
	if !exists {
		return
	}

	filtered := make([]*subscription, 0, len(subs))
	for _, sub := range subs {
		if sub.id == id {
			continue
		}
		filtered = append(filtered, sub)
	}
	b.subs[eventName] = filtered
}

// HasSubscribers 检查指定事件是否有订阅者
func (b *Bus) HasSubscribers(eventName string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	subs, exists := b.subs[eventName]
	return exists && len(subs) > 0
}
