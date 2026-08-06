package discord

import (
	"context"
	"fmt"
	"strings"
	"sync"

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

	return &Bot{
		session: session,
	}
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
