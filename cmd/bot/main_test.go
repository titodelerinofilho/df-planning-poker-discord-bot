package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"
)

func TestRunWritesStartupMessage(t *testing.T) {
	setRequiredEnv(t)

	var stdout bytes.Buffer

	err := run(context.Background(), &stdout)

	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	want := startupMessage + " env=development command_registration_mode=guild\n"
	got := stdout.String()

	if got != want {
		t.Fatalf("run() output = %q, want %q", got, want)
	}
}

func TestRunReturnsConfigurationError(t *testing.T) {
	unsetRequiredEnv(t)

	var stdout bytes.Buffer

	err := run(context.Background(), &stdout)

	if err == nil {
		t.Fatal("run() error = nil, want configuration error")
	}

	if stdout.Len() != 0 {
		t.Fatalf("run() wrote %q, want no output", stdout.String())
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
