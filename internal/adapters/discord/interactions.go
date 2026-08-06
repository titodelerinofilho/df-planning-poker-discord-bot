package discord

import (
	"fmt"

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

	if data.Name != pingCommandName {
		return nil
	}

	latency := bot.session.HeartbeatLatency()
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Pong! Latencia: %dms", latency.Milliseconds()),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}

	err := bot.session.InteractionRespond(event.Interaction, response)

	if err != nil {
		return fmt.Errorf("respond to ping command: %w", err)
	}

	return nil
}
