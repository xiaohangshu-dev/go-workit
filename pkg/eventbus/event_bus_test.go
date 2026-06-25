package eventbus

import (
	"context"
	"errors"
	"testing"
)

type testEvent struct {
	name string
}

func (e testEvent) EventName() string {
	return e.name
}

func TestNewBusIsNotNil(t *testing.T) {
	bus := New()
	if bus == nil {
		t.Fatal("expected New() to return non-nil bus")
	}
}

func TestPublishWithNoSubscribersDoesNotError(t *testing.T) {
	bus := New()
	err := bus.Publish(context.Background(), testEvent{name: "test.event"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestSubscribeAndPublishCallsHandler(t *testing.T) {
	bus := New()
	called := false

	bus.SubscribeFunc("test.event", func(ctx context.Context, event Event) error {
		called = true
		if event.EventName() != "test.event" {
			t.Fatalf("expected event name 'test.event', got %q", event.EventName())
		}
		return nil
	})

	err := bus.Publish(context.Background(), testEvent{name: "test.event"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestSubscribeAndPublishCallsMultipleHandlers(t *testing.T) {
	bus := New()
	calls := 0

	for i := 0; i < 3; i++ {
		bus.SubscribeFunc("test.event", func(ctx context.Context, event Event) error {
			calls++
			return nil
		})
	}

	err := bus.Publish(context.Background(), testEvent{name: "test.event"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 handler calls, got %d", calls)
	}
}

func TestPublishCollectsErrors(t *testing.T) {
	bus := New()
	bus.SubscribeFunc("test.event", func(ctx context.Context, event Event) error {
		return errors.New("handler error")
	})

	err := bus.Publish(context.Background(), testEvent{name: "test.event"})
	if err == nil {
		t.Fatal("expected error from handler")
	}
}

func TestUnsubscribeRemovesHandler(t *testing.T) {
	bus := New()
	called := false

	id := bus.SubscribeFunc("test.event", func(ctx context.Context, event Event) error {
		called = true
		return nil
	})

	bus.Unsubscribe("test.event", id)

	err := bus.Publish(context.Background(), testEvent{name: "test.event"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if called {
		t.Fatal("expected handler not to be called after unsubscribe")
	}
}

func TestUnsubscribeDoesNotAffectOtherHandlers(t *testing.T) {
	bus := New()
	calls := 0

	id1 := bus.SubscribeFunc("test.event", func(ctx context.Context, event Event) error {
		calls++
		return nil
	})
	bus.SubscribeFunc("test.event", func(ctx context.Context, event Event) error {
		calls++
		return nil
	})

	bus.Unsubscribe("test.event", id1)

	err := bus.Publish(context.Background(), testEvent{name: "test.event"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 handler call (other handler), got %d", calls)
	}
}

func TestHasSubscribers(t *testing.T) {
	bus := New()

	if bus.HasSubscribers("test.event") {
		t.Fatal("expected no subscribers initially")
	}

	bus.SubscribeFunc("test.event", func(ctx context.Context, event Event) error {
		return nil
	})

	if !bus.HasSubscribers("test.event") {
		t.Fatal("expected subscribers after subscribe")
	}
}

func TestSubscribeAndPublishWithHandlerInterface(t *testing.T) {
	bus := New()
	called := false

	bus.Subscribe("test.event", HandlerFunc(func(ctx context.Context, event Event) error {
		called = true
		return nil
	}))

	err := bus.Publish(context.Background(), testEvent{name: "test.event"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !called {
		t.Fatal("expected Handler interface to work")
	}
}

func TestSubscribeReturnsUniqueIDs(t *testing.T) {
	bus := New()

	id1 := bus.SubscribeFunc("test.event", func(ctx context.Context, event Event) error { return nil })
	id2 := bus.SubscribeFunc("test.event", func(ctx context.Context, event Event) error { return nil })

	if id1 == id2 {
		t.Fatal("expected unique subscription IDs")
	}
}

func TestSubscribeFuncReturnsNonZeroID(t *testing.T) {
	bus := New()
	id := bus.SubscribeFunc("test.event", func(ctx context.Context, event Event) error { return nil })

	if id == 0 {
		t.Fatal("expected non-zero subscription ID")
	}
}
