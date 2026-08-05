package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/titodelerinofilho/df-planning-poker-discord-bot/internal/platform/config"
	"github.com/titodelerinofilho/df-planning-poker-discord-bot/internal/platform/logging"
	"github.com/titodelerinofilho/df-planning-poker-discord-bot/internal/platform/shutdown"
)

const (
	appVersion               = "dev"
	startupCorrelationID     = "startup"
	shutdownCorrelationID    = "shutdown"
	startupMessage           = "df planning poker bot started"
	shutdownMessage          = "df planning poker bot stopped"
	startupRequestOperation  = "startup.request"
	startupResponseOperation = "startup.response"
	shutdownRequestOperation = "shutdown.request"
	shutdownCloseOperation   = "shutdown.close"
)

func main() {
	ctx, stop := shutdown.Context(context.Background())
	defer stop()

	err := run(ctx, os.Stdout, os.Stderr)

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

	err = shutdown.Wait(ctx)

	if err != nil {
		return err
	}

	handlers.InfoRequest(ctx, shutdownCorrelationID, map[string]string{
		"operation": shutdownRequestOperation,
	})

	err = shutdown.Close(context.Background(), cfg.ShutdownTimeout)

	if err != nil {
		return fmt.Errorf("shutdown resources: %w", err)
	}

	handlers.InfoResponse(ctx, shutdownCorrelationID, map[string]string{
		"operation": shutdownCloseOperation,
		"message":   shutdownMessage,
	})

	return nil
}
