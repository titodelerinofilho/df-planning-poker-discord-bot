package discord

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const botTokenPrefix = "Bot "

type Bot struct {
	session gatewaySession

	mu     sync.Mutex
	opened bool
}

type gatewaySession interface {
	Open() error
	Close() error
	SetIntents(discordgo.Intent)
	AddInteractionCreateHandler(func(*discordgo.InteractionCreate)) func()
	HeartbeatLatency() time.Duration
	InteractionRespond(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
	ApplicationCommands(appID, guildID string, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error)
	ApplicationCommandCreate(appID, guildID string, cmd *discordgo.ApplicationCommand, options ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error)
	ApplicationCommandEdit(appID, guildID, cmdID string, cmd *discordgo.ApplicationCommand, options ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error)
	ApplicationCommandDelete(appID, guildID, cmdID string, options ...discordgo.RequestOption) error
}

func NewBot(token string) (*Bot, error) {
	session, err := discordgo.New(botTokenPrefix + strings.TrimSpace(token))

	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	return newBot(&discordSession{session: session}), nil
}

func (bot *Bot) Open(ctx context.Context) error {
	err := ctx.Err()

	if err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}

	bot.mu.Lock()
	defer bot.mu.Unlock()

	if bot.opened {
		return nil
	}

	err = bot.session.Open()

	if err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}

	bot.opened = true

	return nil
}

func (bot *Bot) Close(ctx context.Context) error {
	err := ctx.Err()

	if err != nil {
		return fmt.Errorf("close discord session: %w", err)
	}

	bot.mu.Lock()
	defer bot.mu.Unlock()

	if !bot.opened {
		return nil
	}

	err = bot.session.Close()

	if err != nil {
		return fmt.Errorf("close discord session: %w", err)
	}

	bot.opened = false

	return nil
}

func newBot(session gatewaySession) *Bot {
	session.SetIntents(discordgo.IntentsGuilds)

	bot := &Bot{
		session: session,
	}

	session.AddInteractionCreateHandler(func(interaction *discordgo.InteractionCreate) {
		_ = bot.handleInteractionCreate(interaction)
	})

	return bot
}

type discordSession struct {
	session *discordgo.Session
}

func (session *discordSession) Open() error {
	return session.session.Open()
}

func (session *discordSession) Close() error {
	return session.session.Close()
}

func (session *discordSession) SetIntents(intents discordgo.Intent) {
	session.session.Identify.Intents = intents
}

func (session *discordSession) AddInteractionCreateHandler(handler func(*discordgo.InteractionCreate)) func() {
	return session.session.AddHandler(func(_ *discordgo.Session, interaction *discordgo.InteractionCreate) {
		handler(interaction)
	})
}

func (session *discordSession) HeartbeatLatency() time.Duration {
	return session.session.HeartbeatLatency()
}

func (session *discordSession) InteractionRespond(interaction *discordgo.Interaction, response *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	return session.session.InteractionRespond(interaction, response, options...)
}

func (session *discordSession) ApplicationCommands(appID, guildID string, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error) {
	return session.session.ApplicationCommands(appID, guildID, options...)
}

func (session *discordSession) ApplicationCommandCreate(appID, guildID string, cmd *discordgo.ApplicationCommand, options ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error) {
	return session.session.ApplicationCommandCreate(appID, guildID, cmd, options...)
}

func (session *discordSession) ApplicationCommandEdit(appID, guildID, cmdID string, cmd *discordgo.ApplicationCommand, options ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error) {
	return session.session.ApplicationCommandEdit(appID, guildID, cmdID, cmd, options...)
}

func (session *discordSession) ApplicationCommandDelete(appID, guildID, cmdID string, options ...discordgo.RequestOption) error {
	return session.session.ApplicationCommandDelete(appID, guildID, cmdID, options...)
}
