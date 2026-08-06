package discord

import (
	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) handleInteractionCreate(event *discordgo.InteractionCreate) error {
	if event == nil || event.Interaction == nil {
		return nil
	}

	if event.Type != discordgo.InteractionApplicationCommand {
		return nil
	}

	data, ok := event.Data.(discordgo.ApplicationCommandInteractionData)

	if !ok {
		return nil
	}

	return bot.router.Route(event, data)
}
