package main

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestRunWritesStartupMessage(t *testing.T) {
	var stdout bytes.Buffer

	err := run(context.Background(), &stdout)

	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	want := startupMessage + "\n"
	got := stdout.String()

	if got != want {
		t.Fatalf("run() output = %q, want %q", got, want)
	}
}

func TestRunReturnsContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var stdout bytes.Buffer

	err := run(ctx, &stdout)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run() error = %v, want context.Canceled", err)
	}

	if stdout.Len() != 0 {
		t.Fatalf("run() wrote %q, want no output", stdout.String())
	}
}
