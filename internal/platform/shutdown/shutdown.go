package shutdown

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var ErrTimeout = errors.New("shutdown timeout")

type Closer func(context.Context) error

func Context(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
}

func Wait(ctx context.Context) error {
	<-ctx.Done()

	return nil
}

func Close(ctx context.Context, timeout time.Duration, closers ...Closer) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for _, closer := range closers {
		err := closer(shutdownCtx)

		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("%w: %s", ErrTimeout, timeout)
			}

			return fmt.Errorf("close resource: %w", err)
		}

		err = shutdownCtx.Err()

		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%w: %s", ErrTimeout, timeout)
		}
	}

	return nil
}
