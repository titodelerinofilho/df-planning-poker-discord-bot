package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) handlePingCommand(event *discordgo.InteractionCreate) error {
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
