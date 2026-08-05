package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	RegistrationModeGuild  RegistrationMode = "guild"
	RegistrationModeGlobal RegistrationMode = "global"
)

var (
	ErrMissingRequired = errors.New("missing required configuration")
	ErrInvalidValue    = errors.New("invalid configuration value")
)

type RegistrationMode string

type Config struct {
	AppEnv                  string
	LogLevel                string
	DiscordApplicationID    string
	DiscordGuildID          string
	CommandRegistrationMode RegistrationMode
	SessionExpiration       time.Duration
	ShutdownTimeout         time.Duration

	discordToken string
	databaseURL  string
}

type lookupEnvFunc func(string) (string, bool)

func LoadFromEnv() (Config, error) {
	return load(os.LookupEnv)
}

func (cfg Config) DiscordToken() string {
	return cfg.discordToken
}

func (cfg Config) DatabaseURL() string {
	return cfg.databaseURL
}

func (cfg Config) String() string {
	return cfg.safeSummary()
}

func (cfg Config) GoString() string {
	return cfg.safeSummary()
}

func load(lookupEnv lookupEnvFunc) (Config, error) {
	values := envValues{
		lookupEnv: lookupEnv,
		missing:   make([]string, 0),
		invalid:   make([]string, 0),
	}

	mode := RegistrationMode(values.required("COMMAND_REGISTRATION_MODE"))
	cfg := Config{
		AppEnv:                  values.required("APP_ENV"),
		LogLevel:                values.required("LOG_LEVEL"),
		DiscordApplicationID:    values.required("DISCORD_APPLICATION_ID"),
		DiscordGuildID:          strings.TrimSpace(values.optional("DISCORD_GUILD_ID")),
		CommandRegistrationMode: mode,
		SessionExpiration:       values.requiredDuration("SESSION_EXPIRATION"),
		ShutdownTimeout:         values.requiredDuration("SHUTDOWN_TIMEOUT"),
		discordToken:            values.required("DISCORD_TOKEN"),
		databaseURL:             values.required("DATABASE_URL"),
	}

	if len(values.missing) > 0 {
		return Config{}, fmt.Errorf("%w: %s", ErrMissingRequired, strings.Join(values.missing, ", "))
	}

	if len(values.invalid) > 0 {
		return Config{}, fmt.Errorf("%w: %s", ErrInvalidValue, strings.Join(values.invalid, ", "))
	}

	err := cfg.validate()

	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) validate() error {
	switch cfg.CommandRegistrationMode {
	case RegistrationModeGuild:
		if cfg.DiscordGuildID == "" {
			return fmt.Errorf("%w: DISCORD_GUILD_ID is required when COMMAND_REGISTRATION_MODE is guild", ErrMissingRequired)
		}
	case RegistrationModeGlobal:
	default:
		return fmt.Errorf("%w: COMMAND_REGISTRATION_MODE must be guild or global", ErrInvalidValue)
	}

	return nil
}

func (cfg Config) safeSummary() string {
	return fmt.Sprintf(
		"Config{AppEnv:%q LogLevel:%q DiscordApplicationID:%q DiscordGuildID:%q CommandRegistrationMode:%q SessionExpiration:%q ShutdownTimeout:%q DiscordToken:%q DatabaseURL:%q}",
		cfg.AppEnv,
		cfg.LogLevel,
		cfg.DiscordApplicationID,
		cfg.DiscordGuildID,
		cfg.CommandRegistrationMode,
		cfg.SessionExpiration,
		cfg.ShutdownTimeout,
		"[redacted]",
		"[redacted]",
	)
}

type envValues struct {
	lookupEnv lookupEnvFunc
	missing   []string
	invalid   []string
}

func (values *envValues) required(name string) string {
	value, ok := values.lookupEnv(name)
	value = strings.TrimSpace(value)

	if !ok || value == "" {
		values.missing = append(values.missing, name)
		return ""
	}

	return value
}

func (values *envValues) optional(name string) string {
	value, ok := values.lookupEnv(name)

	if !ok {
		return ""
	}

	return value
}

func (values *envValues) requiredDuration(name string) time.Duration {
	raw := values.required(name)

	if raw == "" {
		return 0
	}

	duration, err := time.ParseDuration(raw)

	if err != nil {
		values.invalid = append(values.invalid, name+" must be a duration")
		return 0
	}

	return duration
}
