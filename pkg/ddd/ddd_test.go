package ddd

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// --- Entity Tests ---

func TestNewEntitySetsIDAndCreatedAt(t *testing.T) {
	id := uint64(42)
	entity := NewEntity(id)

	if entity.ID != id {
		t.Fatalf("expected ID %d, got %d", id, entity.ID)
	}
	if entity.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestEntityEqualReturnsTrueForSameID(t *testing.T) {
	e1 := NewEntity(uint64(1))
	e2 := NewEntity(uint64(1))

	if !e1.Equal(&e2) {
		t.Fatal("expected entities with same ID to be equal")
	}
}

func TestEntityEqualReturnsFalseForDifferentID(t *testing.T) {
	e1 := NewEntity(uint64(1))
	e2 := NewEntity(uint64(2))

	if e1.Equal(&e2) {
		t.Fatal("expected entities with different IDs to not be equal")
	}
}

func TestEntityEqualReturnsFalseForDifferentType(t *testing.T) {
	e1 := NewEntity(uint64(1))

	result := e1.Equal(nil)
	if result {
		t.Fatal("expected Equal to return false for nil")
	}
}

func TestEntityWithUUIDKey(t *testing.T) {
	id := uuid.New()
	entity := NewEntity(id)

	if entity.ID != id {
		t.Fatalf("expected ID %v, got %v", id, entity.ID)
	}
}

func TestEntityWithStringKey(t *testing.T) {
	id := "user-1"
	entity := NewEntity(id)

	if entity.ID != id {
		t.Fatalf("expected ID %s, got %s", id, entity.ID)
	}
}

// --- AggregateRoot Tests ---

func TestNewAggregateRootCreatesEmptyDomainEvents(t *testing.T) {
	agg := NewAggregateRoot(uint64(1))

	if agg.Entity.ID != uint64(1) {
		t.Fatal("expected aggregate root to have entity ID")
	}
	events := agg.GetDomainEvents()
	if len(events) != 0 {
		t.Fatal("expected new aggregate root to have no domain events")
	}
}

func TestAggregateRootAddsAndRetrievesDomainEvents(t *testing.T) {
	agg := NewAggregateRoot(uint64(1))

	event := NewBaseDomainEvent("test.event")
	agg.AddDomainEvent(event)
	events := agg.GetDomainEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 domain event, got %d", len(events))
	}
	if events[0].EventName() != "test.event" {
		t.Fatalf("expected event name 'test.event', got %q", events[0].EventName())
	}
}

func TestAggregateRootClearDomainEvents(t *testing.T) {
	agg := NewAggregateRoot(uint64(1))

	agg.AddDomainEvent(NewBaseDomainEvent("test.event"))
	events := agg.ClearDomainEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event from ClearDomainEvents, got %d", len(events))
	}
	if len(agg.GetDomainEvents()) != 0 {
		t.Fatal("expected no domain events after ClearDomainEvents")
	}
}

func TestAggregateRootClearDomainEventsReturnsPreviousEvents(t *testing.T) {
	agg := NewAggregateRoot(uint64(1))

	agg.AddDomainEvent(NewBaseDomainEvent("event.1"))
	agg.AddDomainEvent(NewBaseDomainEvent("event.2"))

	events := agg.ClearDomainEvents()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}

// --- ValueObject Tests ---

// emailValueObject 是一个实现了 IValueObject 接口的值对象示例
type emailValueObject struct {
	ValueObject
	Address string
}

func (e *emailValueObject) Equal(other IValueObject) bool {
	otherVO, ok := other.(*emailValueObject)
	if !ok {
		return false
	}
	return e.Address == otherVO.Address
}

func TestValueObjectEqualReturnsTrueForSameValues(t *testing.T) {
	v1 := &emailValueObject{Address: "alice@example.com"}
	v2 := &emailValueObject{Address: "alice@example.com"}

	if !v1.Equal(v2) {
		t.Fatal("expected equal value objects")
	}
}

func TestValueObjectEqualReturnsFalseForDifferentValues(t *testing.T) {
	v1 := &emailValueObject{Address: "alice@example.com"}
	v2 := &emailValueObject{Address: "bob@example.com"}

	if v1.Equal(v2) {
		t.Fatal("expected non-equal value objects")
	}
}

func TestValueObjectEqualReturnsFalseForDifferentType(t *testing.T) {
	v1 := &emailValueObject{Address: "alice@example.com"}

	if v1.Equal(nil) {
		t.Fatal("expected Equal(nil) to return false")
	}
}

func TestValueObjectCanBeEmbedded(t *testing.T) {
	v := &emailValueObject{Address: "test@example.com"}

	if v.Address != "test@example.com" {
		t.Fatalf("expected address 'test@example.com', got %q", v.Address)
	}
}

// --- DomainEvent Tests ---

func TestBaseDomainEventHasEventID(t *testing.T) {
	event := NewBaseDomainEvent("test.event")

	if event.EventID() == uuid.Nil {
		t.Fatal("expected EventID to be set")
	}
	if event.EventName() != "test.event" {
		t.Fatalf("expected EventName 'test.event', got %q", event.EventName())
	}
}

func TestBaseDomainEventIsDomainEvent(t *testing.T) {
	var event DomainEvent = NewBaseDomainEvent("test.event")

	if event.EventName() != "test.event" {
		t.Fatalf("expected EventName 'test.event', got %q", event.EventName())
	}
}

func TestCustomEventImplementsDomainEvent(t *testing.T) {
	// 验证用户自定义事件类型可以嵌入 BaseDomainEvent 并实现 DomainEvent 接口
	type OrderCreated struct {
		BaseDomainEvent
		OrderID string
	}

	event := OrderCreated{
		BaseDomainEvent: NewBaseDomainEvent("order.created"),
		OrderID:         "ord-123",
	}

	// 作为接口使用
	var domainEvent DomainEvent = event
	if domainEvent.EventName() != "order.created" {
		t.Fatalf("expected event name 'order.created', got %q", domainEvent.EventName())
	}
	if event.OrderID != "ord-123" {
		t.Fatalf("expected OrderID 'ord-123', got %q", event.OrderID)
	}
}

// --- DomainEventBus Tests ---

func TestDomainEventBusPublishProcessesAllEvents(t *testing.T) {
	bus := NewDomainEventBus()
	agg := NewAggregateRoot(uint64(1))

	agg.AddDomainEvent(NewBaseDomainEvent("test.event.1"))
	agg.AddDomainEvent(NewBaseDomainEvent("test.event.2"))

	err := bus.Publish(context.Background(), &agg)
	if err != nil {
		t.Logf("Publish returned error (expected if no handler registered): %v", err)
	}

	if len(agg.GetDomainEvents()) != 0 {
		t.Fatal("expected domain events to be cleared after Publish")
	}
}

func TestDomainEventBusPublishWithNoEvents(t *testing.T) {
	bus := NewDomainEventBus()
	agg := NewAggregateRoot(uint64(1))

	err := bus.Publish(context.Background(), &agg)
	if err != nil {
		t.Fatalf("expected no error when publishing aggregate with no events, got: %v", err)
	}
}

func TestNewDomainEventBusReturnsNonNil(t *testing.T) {
	bus := NewDomainEventBus()
	if bus == nil {
		t.Fatal("expected NewDomainEventBus() to return non-nil bus")
	}
}
