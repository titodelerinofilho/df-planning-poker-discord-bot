package discord

import (
	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) handleInteractionCreate(event *discordgo.InteractionCreate) error {
	if event == nil || event.Interaction == nil {
		return nil
	}

	if event.Type == discordgo.InteractionApplicationCommand {
		data, ok := event.Data.(discordgo.ApplicationCommandInteractionData)

		if !ok {
			return nil
		}

		return bot.router.Route(event, data)
	}

	if event.Type == discordgo.InteractionMessageComponent {
		data, ok := event.Data.(discordgo.MessageComponentInteractionData)

		if !ok {
			return nil
		}

		return bot.components.Route(event, data)
	}

	return nil
}
