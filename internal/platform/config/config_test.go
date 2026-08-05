package config

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestLoadReadsTypedConfiguration(t *testing.T) {
	cfg, err := load(lookupEnvFromMap(validEnvironment()))

	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q, want development", cfg.AppEnv)
	}

	if cfg.CommandRegistrationMode != RegistrationModeGuild {
		t.Fatalf("CommandRegistrationMode = %q, want guild", cfg.CommandRegistrationMode)
	}

	if cfg.SessionExpiration != 24*time.Hour {
		t.Fatalf("SessionExpiration = %s, want 24h", cfg.SessionExpiration)
	}

	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("ShutdownTimeout = %s, want 10s", cfg.ShutdownTimeout)
	}

	if cfg.DiscordToken() != "discord-token" {
		t.Fatal("DiscordToken() did not return loaded token")
	}

	if cfg.DatabaseURL() != "postgres://user:password@localhost:5432/df_poker" {
		t.Fatal("DatabaseURL() did not return loaded URL")
	}
}

func TestLoadReportsMissingRequiredVariables(t *testing.T) {
	env := validEnvironment()
	delete(env, "DISCORD_TOKEN")
	delete(env, "DATABASE_URL")

	_, err := load(lookupEnvFromMap(env))

	if !errors.Is(err, ErrMissingRequired) {
		t.Fatalf("load() error = %v, want ErrMissingRequired", err)
	}

	message := err.Error()

	if !strings.Contains(message, "DISCORD_TOKEN") {
		t.Fatalf("load() error = %q, want DISCORD_TOKEN", message)
	}

	if !strings.Contains(message, "DATABASE_URL") {
		t.Fatalf("load() error = %q, want DATABASE_URL", message)
	}
}

func TestLoadRequiresGuildIDForGuildRegistration(t *testing.T) {
	env := validEnvironment()
	delete(env, "DISCORD_GUILD_ID")

	_, err := load(lookupEnvFromMap(env))

	if !errors.Is(err, ErrMissingRequired) {
		t.Fatalf("load() error = %v, want ErrMissingRequired", err)
	}

	if !strings.Contains(err.Error(), "DISCORD_GUILD_ID") {
		t.Fatalf("load() error = %q, want DISCORD_GUILD_ID", err.Error())
	}
}

func TestLoadAllowsMissingGuildIDForGlobalRegistration(t *testing.T) {
	env := validEnvironment()
	env["COMMAND_REGISTRATION_MODE"] = "global"
	delete(env, "DISCORD_GUILD_ID")

	_, err := load(lookupEnvFromMap(env))

	if err != nil {
		t.Fatalf("load() error = %v", err)
	}
}

func TestLoadRejectsInvalidRegistrationMode(t *testing.T) {
	env := validEnvironment()
	env["COMMAND_REGISTRATION_MODE"] = "development"

	_, err := load(lookupEnvFromMap(env))

	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("load() error = %v, want ErrInvalidValue", err)
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	env := validEnvironment()
	env["SESSION_EXPIRATION"] = "tomorrow"

	_, err := load(lookupEnvFromMap(env))

	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("load() error = %v, want ErrInvalidValue", err)
	}
}

func TestConfigFormattingRedactsSecrets(t *testing.T) {
	cfg, err := load(lookupEnvFromMap(validEnvironment()))

	if err != nil {
		t.Fatalf("load() error = %v", err)
	}

	rendered := fmt.Sprintf("%#v %s", cfg, cfg)

	if strings.Contains(rendered, cfg.DiscordToken()) {
		t.Fatalf("formatted config exposed Discord token: %s", rendered)
	}

	if strings.Contains(rendered, cfg.DatabaseURL()) {
		t.Fatalf("formatted config exposed database URL: %s", rendered)
	}

	if !strings.Contains(rendered, "[redacted]") {
		t.Fatalf("formatted config = %s, want redaction marker", rendered)
	}
}

func validEnvironment() map[string]string {
	return map[string]string{
		"APP_ENV":                   "development",
		"LOG_LEVEL":                 "debug",
		"DISCORD_TOKEN":             "discord-token",
		"DISCORD_APPLICATION_ID":    "discord-application-id",
		"DISCORD_GUILD_ID":          "discord-guild-id",
		"DATABASE_URL":              "postgres://user:password@localhost:5432/df_poker",
		"COMMAND_REGISTRATION_MODE": "guild",
		"SESSION_EXPIRATION":        "24h",
		"SHUTDOWN_TIMEOUT":          "10s",
	}
}

func lookupEnvFromMap(env map[string]string) lookupEnvFunc {
	return func(name string) (string, bool) {
		value, ok := env[name]

		return value, ok
	}
}
