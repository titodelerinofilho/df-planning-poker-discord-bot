package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const pingCommandName = "ping"

type CommandDefinition struct {
	Name        string
	Description string
}

func DevelopmentCommands() []CommandDefinition {
	return []CommandDefinition{
		{
			Name:        pingCommandName,
			Description: "Mostra a latencia basica do bot",
		},
	}
}

func ManagedCommandNames(commands []CommandDefinition) []string {
	names := make([]string, 0, len(commands))

	for _, command := range commands {
		names = append(names, command.Name)
	}

	return names
}

func (bot *Bot) SyncGuildCommands(ctx context.Context, applicationID string, guildID string, commands []CommandDefinition) error {
	guildID = strings.TrimSpace(guildID)

	if guildID == "" {
		return fmt.Errorf("sync guild commands: guild id is required")
	}

	return bot.syncCommands(ctx, "guild", applicationID, guildID, commands, ManagedCommandNames(commands))
}

func (bot *Bot) syncGuildCommands(ctx context.Context, applicationID string, guildID string, commands []CommandDefinition, managedNames []string) error {
	return bot.syncCommands(ctx, "guild", applicationID, guildID, commands, managedNames)
}

func (bot *Bot) SyncGlobalCommands(ctx context.Context, applicationID string, commands []CommandDefinition) error {
	return bot.syncCommands(ctx, "global", applicationID, "", commands, ManagedCommandNames(commands))
}

func (bot *Bot) syncCommands(ctx context.Context, scope string, applicationID string, guildID string, commands []CommandDefinition, managedNames []string) error {
	err := ctx.Err()

	if err != nil {
		return fmt.Errorf("sync %s commands: %w", scope, err)
	}

	applicationID = strings.TrimSpace(applicationID)
	guildID = strings.TrimSpace(guildID)

	if applicationID == "" {
		return fmt.Errorf("sync %s commands: application id is required", scope)
	}

	desired, err := commandMap(commands)

	if err != nil {
		return err
	}

	managed, err := commandNameSet(managedNames)

	if err != nil {
		return err
	}

	existingCommands, err := bot.session.ApplicationCommands(applicationID, guildID)

	if err != nil {
		return fmt.Errorf("list %s commands: %w", scope, err)
	}

	existing := make(map[string]*discordgo.ApplicationCommand, len(existingCommands))

	for _, command := range existingCommands {
		if command == nil {
			continue
		}

		existing[command.Name] = command
	}

	for _, command := range desired {
		discordCommand := command.discordgoCommand()
		existingCommand, exists := existing[command.Name]

		if !exists {
			_, err = bot.session.ApplicationCommandCreate(applicationID, guildID, discordCommand)

			if err != nil {
				return fmt.Errorf("create %s command %q: %w", scope, command.Name, err)
			}

			continue
		}

		if commandMatches(existingCommand, command) {
			continue
		}

		_, err = bot.session.ApplicationCommandEdit(applicationID, guildID, existingCommand.ID, discordCommand)

		if err != nil {
			return fmt.Errorf("update %s command %q: %w", scope, command.Name, err)
		}
	}

	for name, existingCommand := range existing {
		_, wanted := desired[name]
		_, controlled := managed[name]

		if wanted || !controlled {
			continue
		}

		err = bot.session.ApplicationCommandDelete(applicationID, guildID, existingCommand.ID)

		if err != nil {
			return fmt.Errorf("delete %s command %q: %w", scope, name, err)
		}
	}

	return nil
}

func commandMap(commands []CommandDefinition) (map[string]CommandDefinition, error) {
	result := make(map[string]CommandDefinition, len(commands))

	for _, command := range commands {
		command.Name = strings.TrimSpace(command.Name)
		command.Description = strings.TrimSpace(command.Description)

		if command.Name == "" {
			return nil, fmt.Errorf("sync guild commands: command name is required")
		}

		if command.Description == "" {
			return nil, fmt.Errorf("sync guild commands: command description is required")
		}

		_, exists := result[command.Name]

		if exists {
			return nil, fmt.Errorf("sync guild commands: duplicate command %q", command.Name)
		}

		result[command.Name] = command
	}

	return result, nil
}

func commandNameSet(names []string) (map[string]struct{}, error) {
	result := make(map[string]struct{}, len(names))

	for _, name := range names {
		name = strings.TrimSpace(name)

		if name == "" {
			return nil, fmt.Errorf("sync guild commands: managed command name is required")
		}

		result[name] = struct{}{}
	}

	return result, nil
}

func (command CommandDefinition) discordgoCommand() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        strings.TrimSpace(command.Name),
		Description: strings.TrimSpace(command.Description),
	}
}

func commandMatches(existing *discordgo.ApplicationCommand, desired CommandDefinition) bool {
	if existing == nil {
		return false
	}

	return existing.Type == discordgo.ChatApplicationCommand &&
		existing.Name == strings.TrimSpace(desired.Name) &&
		existing.Description == strings.TrimSpace(desired.Description)
}
