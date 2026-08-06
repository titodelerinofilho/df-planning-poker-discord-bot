package discord

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var ErrCommandRouteNotFound = errors.New("discord command route not found")

type commandHandler func(*discordgo.InteractionCreate) error

type commandRouter struct {
	handlers map[string]commandHandler
}

type commandRoute struct {
	Name        string
	Subcommands []string
}

func newDefaultCommandRouter(bot *Bot) *commandRouter {
	router := newCommandRouter()
	router.Handle(commandRoute{Name: pingCommandName}, bot.handlePingCommand)

	return router
}

func newCommandRouter() *commandRouter {
	return &commandRouter{
		handlers: make(map[string]commandHandler),
	}
}

func (router *commandRouter) Handle(route commandRoute, handler commandHandler) {
	router.handlers[route.key()] = handler
}

func (router *commandRouter) Route(event *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) error {
	route := commandRoute{
		Name:        data.Name,
		Subcommands: commandSubcommands(data.Options),
	}

	handler, ok := router.handlers[route.key()]

	if !ok {
		return fmt.Errorf("%w: %s", ErrCommandRouteNotFound, route.String())
	}

	err := handler(event)

	if err != nil {
		return fmt.Errorf("handle command route %s: %w", route.String(), err)
	}

	return nil
}

func (route commandRoute) String() string {
	parts := append([]string{route.Name}, route.Subcommands...)

	return strings.Join(parts, " ")
}

func (route commandRoute) key() string {
	parts := append([]string{strings.TrimSpace(route.Name)}, route.Subcommands...)

	for index, part := range parts {
		parts[index] = strings.TrimSpace(part)
	}

	return strings.Join(parts, "\x00")
}

func commandSubcommands(options []*discordgo.ApplicationCommandInteractionDataOption) []string {
	subcommands := make([]string, 0, len(options))

	for _, option := range options {
		if option == nil {
			continue
		}

		if option.Type != discordgo.ApplicationCommandOptionSubCommand &&
			option.Type != discordgo.ApplicationCommandOptionSubCommandGroup {
			continue
		}

		subcommands = append(subcommands, option.Name)
		subcommands = append(subcommands, commandSubcommands(option.Options)...)
	}

	return subcommands
}
