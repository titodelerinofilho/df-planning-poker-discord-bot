package shutdown

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitReturnsWhenContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Wait(ctx)

	if err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
}

func TestCloseRunsClosers(t *testing.T) {
	var called bool

	err := Close(context.Background(), time.Second, func(context.Context) error {
		called = true

		return nil
	})

	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if !called {
		t.Fatal("Close() did not call closer")
	}
}

func TestCloseReturnsCloserError(t *testing.T) {
	closerErr := errors.New("close discord")

	err := Close(context.Background(), time.Second, func(context.Context) error {
		return closerErr
	})

	if !errors.Is(err, closerErr) {
		t.Fatalf("Close() error = %v, want closer error", err)
	}
}

func TestCloseReturnsTimeout(t *testing.T) {
	err := Close(context.Background(), time.Nanosecond, func(ctx context.Context) error {
		<-ctx.Done()

		return nil
	})

	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Close() error = %v, want ErrTimeout", err)
	}
}

func TestCloseMapsDeadlineExceededToTimeout(t *testing.T) {
	err := Close(context.Background(), time.Second, func(context.Context) error {
		return context.DeadlineExceeded
	})

	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Close() error = %v, want ErrTimeout", err)
	}
}
