package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/titodelerinofilho/df-planning-poker-discord-bot/internal/platform/config"
	"github.com/titodelerinofilho/df-planning-poker-discord-bot/internal/platform/logging"
)

const (
	appVersion               = "dev"
	startupCorrelationID     = "startup"
	startupMessage           = "df planning poker bot started"
	startupRequestOperation  = "startup.request"
	startupResponseOperation = "startup.response"
)

func main() {
	err := run(context.Background(), os.Stdout, os.Stderr)

	if err != nil {
		handlers := logging.NewBootstrapHandlers(os.Stderr, appVersion)
		logging.LogStartupError(context.Background(), handlers, startupCorrelationID, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, stdout io.Writer, stderr io.Writer) error {
	err := ctx.Err()

	if err != nil {
		return fmt.Errorf("start bot: %w", err)
	}

	cfg, err := config.LoadFromEnv()

	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	handlers, err := logging.NewHandlers(logging.Config{
		Environment: cfg.AppEnv,
		Level:       cfg.LogLevel,
		Version:     appVersion,
		Stdout:      stdout,
		Stderr:      stderr,
	})

	if err != nil {
		return fmt.Errorf("configure logging: %w", err)
	}

	handlers.InfoRequest(ctx, startupCorrelationID, map[string]string{
		"operation": startupRequestOperation,
	})

	handlers.InfoResponse(ctx, startupCorrelationID, map[string]string{
		"operation":                 startupResponseOperation,
		"message":                   startupMessage,
		"command_registration_mode": string(cfg.CommandRegistrationMode),
	})

	return nil
}
