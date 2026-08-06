package discord

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestCommandRouterRoutesCommandByName(t *testing.T) {
	router := newCommandRouter()
	calls := 0
	router.Handle(commandRoute{Name: "ping"}, func(*discordgo.InteractionCreate) error {
		calls++

		return nil
	})

	err := router.Route(&discordgo.InteractionCreate{}, discordgo.ApplicationCommandInteractionData{
		Name: "ping",
	})

	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1", calls)
	}
}

func TestCommandRouterRoutesCommandBySubcommand(t *testing.T) {
	router := newCommandRouter()
	calls := 0
	router.Handle(commandRoute{Name: "planning", Subcommands: []string{"start"}}, func(*discordgo.InteractionCreate) error {
		calls++

		return nil
	})

	err := router.Route(&discordgo.InteractionCreate{}, discordgo.ApplicationCommandInteractionData{
		Name: "planning",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Type: discordgo.ApplicationCommandOptionSubCommand,
				Name: "start",
			},
		},
	})

	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1", calls)
	}
}

func TestCommandRouterRoutesCommandBySubcommandGroup(t *testing.T) {
	router := newCommandRouter()
	calls := 0
	router.Handle(commandRoute{Name: "planning", Subcommands: []string{"session", "start"}}, func(*discordgo.InteractionCreate) error {
		calls++

		return nil
	})

	err := router.Route(&discordgo.InteractionCreate{}, discordgo.ApplicationCommandInteractionData{
		Name: "planning",
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Type: discordgo.ApplicationCommandOptionSubCommandGroup,
				Name: "session",
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{
						Type: discordgo.ApplicationCommandOptionSubCommand,
						Name: "start",
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1", calls)
	}
}

func TestCommandRouterReturnsStandardRouteError(t *testing.T) {
	router := newCommandRouter()

	err := router.Route(&discordgo.InteractionCreate{}, discordgo.ApplicationCommandInteractionData{
		Name: "unknown",
	})

	if !errors.Is(err, ErrCommandRouteNotFound) {
		t.Fatalf("Route() error = %v, want ErrCommandRouteNotFound", err)
	}
}

func TestCommandRouterWrapsHandlerError(t *testing.T) {
	handlerErr := errors.New("handler failed")
	router := newCommandRouter()
	router.Handle(commandRoute{Name: "ping"}, func(*discordgo.InteractionCreate) error {
		return handlerErr
	})

	err := router.Route(&discordgo.InteractionCreate{}, discordgo.ApplicationCommandInteractionData{
		Name: "ping",
	})

	if !errors.Is(err, handlerErr) {
		t.Fatalf("Route() error = %v, want handler error", err)
	}
}
