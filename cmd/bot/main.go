package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/titodelerinofilho/df-planning-poker-discord-bot/internal/platform/config"
)

const startupMessage = "df planning poker bot started"

func main() {
	err := run(context.Background(), os.Stdout)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, stdout io.Writer) error {
	err := ctx.Err()

	if err != nil {
		return fmt.Errorf("start bot: %w", err)
	}

	cfg, err := config.LoadFromEnv()

	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	_, err = fmt.Fprintf(stdout, "%s env=%s command_registration_mode=%s\n", startupMessage, cfg.AppEnv, cfg.CommandRegistrationMode)

	if err != nil {
		return fmt.Errorf("write startup message: %w", err)
	}

	return nil
}
