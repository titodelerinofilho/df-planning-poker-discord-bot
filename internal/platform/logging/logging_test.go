package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestHandlersRouteExceptionsToStderrAsCritical(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	handlers, err := NewHandlers(Config{
		Environment: "development",
		Level:       "debug",
		Version:     "test-version",
		Stdout:      &stdout,
		Stderr:      &stderr,
	})

	if err != nil {
		t.Fatalf("NewHandlers() error = %v", err)
	}

	handlers.CriticalException(context.Background(), "3ed68ca9-54ce-4b8d-b02e-5150f23c0cb4", map[string]string{
		"error": "boom",
	})

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want no output", stdout.String())
	}

	entry := decodeLogEntry(t, stderr.String())

	assertLogField(t, entry, "channel", "exceptions")
	assertLogField(t, entry, "level", "CRITICAL")
	assertLogField(t, entry, "message", "exception")
	assertLogField(t, entry, "correlation_identifier", "3ed68ca9-54ce-4b8d-b02e-5150f23c0cb4")
	assertLogField(t, entry, "version", "test-version")
	assertLogField(t, entry, "environment", "development")
	assertContentField(t, entry, "error", "boom")
}

func TestHandlersRouteOperationalLogsToStdoutAsInfo(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	handlers, err := NewHandlers(Config{
		Environment: "development",
		Level:       "info",
		Version:     "test-version",
		Stdout:      &stdout,
		Stderr:      &stderr,
	})

	if err != nil {
		t.Fatalf("NewHandlers() error = %v", err)
	}

	handlers.InfoRequest(context.Background(), "request-id", map[string]string{"method": "POST"})
	handlers.InfoResponse(context.Background(), "response-id", map[string]int{"status": 200})
	handlers.InfoDatabaseQuery(context.Background(), "query-id", map[string]string{"query": "select planning session"})

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want no output", stderr.String())
	}

	entries := decodeLogEntries(t, stdout.String())

	if len(entries) != 3 {
		t.Fatalf("log entries = %d, want 3", len(entries))
	}

	assertLogField(t, entries[0], "channel", "requests")
	assertLogField(t, entries[0], "level", "INFO")
	assertLogField(t, entries[0], "message", "request")
	assertLogField(t, entries[0], "correlation_identifier", "request-id")

	assertLogField(t, entries[1], "channel", "responses")
	assertLogField(t, entries[1], "level", "INFO")
	assertLogField(t, entries[1], "message", "response")
	assertLogField(t, entries[1], "correlation_identifier", "response-id")

	assertLogField(t, entries[2], "channel", "database_queries")
	assertLogField(t, entries[2], "level", "INFO")
	assertLogField(t, entries[2], "message", "database_query")
	assertLogField(t, entries[2], "correlation_identifier", "query-id")
}

func TestHandlersFilterByLevel(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	handlers, err := NewHandlers(Config{
		Environment: "development",
		Level:       "warn",
		Version:     "test-version",
		Stdout:      &stdout,
		Stderr:      &stderr,
	})

	if err != nil {
		t.Fatalf("NewHandlers() error = %v", err)
	}

	handlers.InfoRequest(context.Background(), "request-id", map[string]string{"method": "POST"})

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want no output", stdout.String())
	}
}

func TestNewHandlersRejectsInvalidLevel(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	_, err := NewHandlers(Config{
		Environment: "development",
		Level:       "verbose",
		Version:     "test-version",
		Stdout:      &stdout,
		Stderr:      &stderr,
	})

	if err == nil {
		t.Fatal("NewHandlers() error = nil, want invalid level error")
	}
}

func TestBootstrapHandlersUseUnknownEnvironment(t *testing.T) {
	var stderr bytes.Buffer

	handlers := NewBootstrapHandlers(&stderr, "test-version")

	LogStartupError(context.Background(), handlers, "startup", errTest)

	entry := decodeLogEntry(t, stderr.String())

	assertLogField(t, entry, "channel", "exceptions")
	assertLogField(t, entry, "level", "CRITICAL")
	assertLogField(t, entry, "environment", "unknown")
	assertLogField(t, entry, "correlation_identifier", "startup")
	assertContentField(t, entry, "operation", "startup")
	assertContentField(t, entry, "error", "test error")
}

var errTest = testError{}

type testError struct{}

func (testError) Error() string {
	return "test error"
}

func decodeLogEntry(t *testing.T, raw string) map[string]any {
	t.Helper()

	entries := decodeLogEntries(t, raw)

	if len(entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(entries))
	}

	return entries[0]
}

func decodeLogEntries(t *testing.T, raw string) []map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(raw), "\n")
	entries := make([]map[string]any, 0, len(lines))

	for _, line := range lines {
		var entry map[string]any

		err := json.Unmarshal([]byte(line), &entry)

		if err != nil {
			t.Fatalf("decode log entry %q: %v", line, err)
		}

		entries = append(entries, entry)
	}

	return entries
}

func assertLogField(t *testing.T, entry map[string]any, name string, want string) {
	t.Helper()

	if entry[name] != want {
		t.Fatalf("%s = %v, want %s", name, entry[name], want)
	}
}

func assertContentField(t *testing.T, entry map[string]any, name string, want string) {
	t.Helper()

	content, ok := entry["content"].(map[string]any)

	if !ok {
		t.Fatalf("content = %T, want object", entry["content"])
	}

	if content[name] != want {
		t.Fatalf("content.%s = %v, want %s", name, content[name], want)
	}
}
