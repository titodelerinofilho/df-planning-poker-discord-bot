package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	discordadapter "github.com/titodelerinofilho/df-planning-poker-discord-bot/internal/adapters/discord"
	"github.com/titodelerinofilho/df-planning-poker-discord-bot/internal/platform/config"
	"github.com/titodelerinofilho/df-planning-poker-discord-bot/internal/platform/logging"
	"github.com/titodelerinofilho/df-planning-poker-discord-bot/internal/platform/shutdown"
)

const (
	appVersion               = "dev"
	startupCorrelationID     = "startup"
	shutdownCorrelationID    = "shutdown"
	startupMessage           = "df planning poker bot started"
	discordOpenOperation     = "discord.open"
	discordCloseOperation    = "discord.close"
	guildCommandsOperation   = "discord.guild_commands.sync"
	globalCommandsOperation  = "discord.global_commands.sync"
	shutdownMessage          = "df planning poker bot stopped"
	startupRequestOperation  = "startup.request"
	startupResponseOperation = "startup.response"
	shutdownRequestOperation = "shutdown.request"
	shutdownCloseOperation   = "shutdown.close"
)

func main() {
	ctx, stop := shutdown.Context(context.Background())
	defer stop()

	err := run(ctx, os.Stdout, os.Stderr, defaultRuntimeDependencies())

	if err != nil {
		handlers := logging.NewBootstrapHandlers(os.Stderr, appVersion)
		logging.LogStartupError(context.Background(), handlers, startupCorrelationID, err)
		os.Exit(1)
	}
}

type discordBot interface {
	Open(context.Context) error
	Close(context.Context) error
	SyncGuildCommands(context.Context, string, string, []discordadapter.CommandDefinition) error
	SyncGlobalCommands(context.Context, string, []discordadapter.CommandDefinition) error
}

type runtimeDependencies struct {
	newDiscordBot func(string) (discordBot, error)
}

func defaultRuntimeDependencies() runtimeDependencies {
	return runtimeDependencies{
		newDiscordBot: func(token string) (discordBot, error) {
			return discordadapter.NewBot(token)
		},
	}
}

func run(ctx context.Context, stdout io.Writer, stderr io.Writer, dependencies runtimeDependencies) error {
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

	discordBot, err := dependencies.newDiscordBot(cfg.DiscordToken())

	if err != nil {
		return fmt.Errorf("create discord bot: %w", err)
	}

	handlers.InfoRequest(ctx, startupCorrelationID, map[string]string{
		"operation": discordOpenOperation,
	})

	err = discordBot.Open(ctx)

	if err != nil {
		return fmt.Errorf("open discord bot: %w", err)
	}

	handlers.InfoResponse(ctx, startupCorrelationID, map[string]string{
		"operation": discordOpenOperation,
	})

	if cfg.CommandRegistrationMode == config.RegistrationModeGuild {
		commands := discordadapter.DevelopmentCommands()

		handlers.InfoRequest(ctx, startupCorrelationID, map[string]string{
			"operation":      guildCommandsOperation,
			"application_id": cfg.DiscordApplicationID,
			"guild_id":       cfg.DiscordGuildID,
		})

		err = discordBot.SyncGuildCommands(ctx, cfg.DiscordApplicationID, cfg.DiscordGuildID, commands)

		if err != nil {
			closeErr := shutdown.Close(context.Background(), cfg.ShutdownTimeout, discordBot.Close)

			if closeErr != nil {
				return fmt.Errorf("sync guild commands: %w", errors.Join(err, fmt.Errorf("close discord bot: %w", closeErr)))
			}

			return fmt.Errorf("sync guild commands: %w", err)
		}

		handlers.InfoResponse(ctx, startupCorrelationID, map[string]string{
			"operation":      guildCommandsOperation,
			"application_id": cfg.DiscordApplicationID,
			"guild_id":       cfg.DiscordGuildID,
		})
	}

	if cfg.CommandRegistrationMode == config.RegistrationModeGlobal {
		commands := discordadapter.DevelopmentCommands()

		handlers.InfoRequest(ctx, startupCorrelationID, map[string]string{
			"operation":      globalCommandsOperation,
			"application_id": cfg.DiscordApplicationID,
		})

		err = discordBot.SyncGlobalCommands(ctx, cfg.DiscordApplicationID, commands)

		if err != nil {
			closeErr := shutdown.Close(context.Background(), cfg.ShutdownTimeout, discordBot.Close)

			if closeErr != nil {
				return fmt.Errorf("sync global commands: %w", errors.Join(err, fmt.Errorf("close discord bot: %w", closeErr)))
			}

			return fmt.Errorf("sync global commands: %w", err)
		}

		handlers.InfoResponse(ctx, startupCorrelationID, map[string]string{
			"operation":      globalCommandsOperation,
			"application_id": cfg.DiscordApplicationID,
		})
	}

	err = shutdown.Wait(ctx)

	if err != nil {
		return err
	}

	handlers.InfoRequest(ctx, shutdownCorrelationID, map[string]string{
		"operation": shutdownRequestOperation,
	})

	err = shutdown.Close(context.Background(), cfg.ShutdownTimeout, discordBot.Close)

	if err != nil {
		return fmt.Errorf("shutdown resources: %w", err)
	}

	handlers.InfoResponse(ctx, shutdownCorrelationID, map[string]string{
		"operation": discordCloseOperation,
	})

	handlers.InfoResponse(ctx, shutdownCorrelationID, map[string]string{
		"operation": shutdownCloseOperation,
		"message":   shutdownMessage,
	})

	return nil
}
