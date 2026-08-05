package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunLogsStartupMessage(t *testing.T) {
	setRequiredEnv(t)

	stdout := newLogFile(t)
	stderr := newLogFile(t)

	err := run(context.Background(), stdout, stderr)

	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	stderrOutput := readLogFile(t, stderr)

	if stderrOutput != "" {
		t.Fatalf("stderr = %q, want no output", stderrOutput)
	}

	stdoutOutput := readLogFile(t, stdout)
	entries := decodeMainLogEntries(t, stdoutOutput)

	if len(entries) != 2 {
		t.Fatalf("log entries = %d, want 2", len(entries))
	}

	if entries[0]["channel"] != "requests" {
		t.Fatalf("first channel = %v, want requests", entries[0]["channel"])
	}

	if entries[0]["level"] != "INFO" {
		t.Fatalf("first level = %v, want INFO", entries[0]["level"])
	}

	if entries[0]["correlation_identifier"] != startupCorrelationID {
		t.Fatalf("first correlation_identifier = %v, want %s", entries[0]["correlation_identifier"], startupCorrelationID)
	}

	if entries[1]["channel"] != "responses" {
		t.Fatalf("second channel = %v, want responses", entries[1]["channel"])
	}

	if entries[1]["level"] != "INFO" {
		t.Fatalf("second level = %v, want INFO", entries[1]["level"])
	}

	if entries[1]["environment"] != "development" {
		t.Fatalf("environment = %v, want development", entries[1]["environment"])
	}

	if entries[1]["version"] != appVersion {
		t.Fatalf("version = %v, want %s", entries[1]["version"], appVersion)
	}

	if entries[1]["correlation_identifier"] != startupCorrelationID {
		t.Fatalf("second correlation_identifier = %v, want %s", entries[1]["correlation_identifier"], startupCorrelationID)
	}

	content, ok := entries[1]["content"].(map[string]any)

	if !ok {
		t.Fatalf("content = %T, want object", entries[1]["content"])
	}

	if content["message"] != startupMessage {
		t.Fatalf("content.message = %v, want %s", content["message"], startupMessage)
	}

	if content["command_registration_mode"] != "guild" {
		t.Fatalf("content.command_registration_mode = %v, want guild", content["command_registration_mode"])
	}
}

func TestRunReturnsConfigurationError(t *testing.T) {
	unsetRequiredEnv(t)

	stdout := newLogFile(t)
	stderr := newLogFile(t)

	err := run(context.Background(), stdout, stderr)

	if err == nil {
		t.Fatal("run() error = nil, want configuration error")
	}

	stdoutOutput := readLogFile(t, stdout)

	if stdoutOutput != "" {
		t.Fatalf("run() wrote stdout %q, want no output", stdoutOutput)
	}

	stderrOutput := readLogFile(t, stderr)

	if stderrOutput != "" {
		t.Fatalf("run() wrote stderr %q, want no output", stderrOutput)
	}
}

func TestRunReturnsContextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stdout := newLogFile(t)
	stderr := newLogFile(t)

	err := run(ctx, stdout, stderr)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run() error = %v, want context.Canceled", err)
	}

	stdoutOutput := readLogFile(t, stdout)

	if stdoutOutput != "" {
		t.Fatalf("run() wrote stdout %q, want no output", stdoutOutput)
	}

	stderrOutput := readLogFile(t, stderr)

	if stderrOutput != "" {
		t.Fatalf("run() wrote stderr %q, want no output", stderrOutput)
	}
}

func TestRunRejectsInvalidLogLevel(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("LOG_LEVEL", "verbose")

	stdout := newLogFile(t)
	stderr := newLogFile(t)

	err := run(context.Background(), stdout, stderr)

	if err == nil {
		t.Fatal("run() error = nil, want invalid log level error")
	}

	if !strings.Contains(err.Error(), "configure logging") {
		t.Fatalf("run() error = %v, want configure logging context", err)
	}
}

func setRequiredEnv(t *testing.T) {
	t.Helper()

	t.Setenv("APP_ENV", "development")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DISCORD_TOKEN", "discord-token")
	t.Setenv("DISCORD_APPLICATION_ID", "discord-application-id")
	t.Setenv("DISCORD_GUILD_ID", "discord-guild-id")
	t.Setenv("DATABASE_URL", "postgres://user:password@localhost:5432/df_poker")
	t.Setenv("COMMAND_REGISTRATION_MODE", "guild")
	t.Setenv("SESSION_EXPIRATION", "24h")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
}

func unsetRequiredEnv(t *testing.T) {
	t.Helper()

	names := []string{
		"APP_ENV",
		"LOG_LEVEL",
		"DISCORD_TOKEN",
		"DISCORD_APPLICATION_ID",
		"DISCORD_GUILD_ID",
		"DATABASE_URL",
		"COMMAND_REGISTRATION_MODE",
		"SESSION_EXPIRATION",
		"SHUTDOWN_TIMEOUT",
	}

	original := make(map[string]string, len(names))
	present := make(map[string]bool, len(names))

	for _, name := range names {
		value, ok := os.LookupEnv(name)
		original[name] = value
		present[name] = ok

		err := os.Unsetenv(name)

		if err != nil {
			t.Fatalf("unset %s: %v", name, err)
		}
	}

	t.Cleanup(func() {
		for _, name := range names {
			if !present[name] {
				os.Unsetenv(name)
				continue
			}

			os.Setenv(name, original[name])
		}
	})
}

func newLogFile(t *testing.T) *os.File {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "bot-log-*.json")

	if err != nil {
		t.Fatalf("create temp log file: %v", err)
	}

	return file
}

func readLogFile(t *testing.T, file *os.File) string {
	t.Helper()

	_, err := file.Seek(0, io.SeekStart)

	if err != nil {
		t.Fatalf("seek log file: %v", err)
	}

	content, err := io.ReadAll(file)

	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	return string(content)
}

func decodeMainLogEntries(t *testing.T, raw string) []map[string]any {
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
