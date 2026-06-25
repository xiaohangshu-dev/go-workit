package ddd

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent 领域事件接口，所有领域事件需实现此接口。
type DomainEvent interface {
	// EventID 返回事件的唯一标识符。
	EventID() uuid.UUID
	// CreatedAt 返回事件的创建时间。
	CreatedAt() time.Time
	// EventName 返回事件名称，用于事件路由和处理器匹配。
	EventName() string
}

// BaseDomainEvent 基础领域事件，实现了 DomainEvent 接口。
// 用户定义自定义事件时嵌入此结构体即可。
//
// 使用方式:
//
//	type OrderCreated struct {
//	    ddd.BaseDomainEvent
//	    OrderID string
//	}
//
//	event := OrderCreated{
//	    BaseDomainEvent: ddd.NewBaseDomainEvent("order.created"),
//	    OrderID:         "ord-123",
//	}
type BaseDomainEvent struct {
	eventID   uuid.UUID
	created   time.Time
	eventName string
}

// NewBaseDomainEvent 创建一个新的基础领域事件。
func NewBaseDomainEvent(eventName string) BaseDomainEvent {
	return BaseDomainEvent{
		eventID:   uuid.New(),
		created:   time.Now(),
		eventName: eventName,
	}
}

// EventID 返回事件的唯一标识符。
func (e BaseDomainEvent) EventID() uuid.UUID {
	return e.eventID
}

// CreatedAt 返回事件的创建时间。
func (e BaseDomainEvent) CreatedAt() time.Time {
	return e.created
}

// EventName 返回事件名称。
func (e BaseDomainEvent) EventName() string {
	return e.eventName
}
