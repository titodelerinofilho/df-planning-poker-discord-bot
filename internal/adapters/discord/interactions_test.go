package discord

import (
	"errors"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestHandleInteractionCreateRespondsToPingEphemerally(t *testing.T) {
	session := &fakeSession{
		heartbeatLatency: 42 * time.Millisecond,
	}
	bot := newBot(session)

	err := bot.handleInteractionCreate(pingInteraction())

	if err != nil {
		t.Fatalf("handleInteractionCreate() error = %v", err)
	}

	if session.interactionRespondCalls != 1 {
		t.Fatalf("interaction respond calls = %d, want 1", session.interactionRespondCalls)
	}

	response := session.interactionResponse

	if response == nil {
		t.Fatal("interaction response is nil")
	}

	if response.Type != discordgo.InteractionResponseChannelMessageWithSource {
		t.Fatalf("response type = %v, want channel message", response.Type)
	}

	if response.Data == nil {
		t.Fatal("response data is nil")
	}

	if response.Data.Content != "Pong! Latencia: 42ms" {
		t.Fatalf("response content = %q, want ping latency", response.Data.Content)
	}

	if response.Data.Flags != discordgo.MessageFlagsEphemeral {
		t.Fatalf("response flags = %v, want ephemeral", response.Data.Flags)
	}
}

func TestRegisteredInteractionHandlerHandlesPing(t *testing.T) {
	session := &fakeSession{
		heartbeatLatency: 10 * time.Millisecond,
	}

	newBot(session)

	if session.interactionCreateHandler == nil {
		t.Fatal("interaction handler was not registered")
	}

	session.interactionCreateHandler(pingInteraction())

	if session.interactionRespondCalls != 1 {
		t.Fatalf("interaction respond calls = %d, want 1", session.interactionRespondCalls)
	}
}

func TestHandleInteractionCreateIgnoresUnknownCommand(t *testing.T) {
	session := &fakeSession{}
	bot := newBot(session)
	event := applicationCommandInteraction("status")

	err := bot.handleInteractionCreate(event)

	if !errors.Is(err, ErrCommandRouteNotFound) {
		t.Fatalf("handleInteractionCreate() error = %v, want route not found", err)
	}

	if session.interactionRespondCalls != 0 {
		t.Fatalf("interaction respond calls = %d, want 0", session.interactionRespondCalls)
	}
}

func TestHandleInteractionCreateIgnoresNonCommandInteraction(t *testing.T) {
	session := &fakeSession{}
	bot := newBot(session)
	event := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:    "interaction-id",
			Token: "interaction-token",
			Type:  discordgo.InteractionMessageComponent,
		},
	}

	err := bot.handleInteractionCreate(event)

	if err != nil {
		t.Fatalf("handleInteractionCreate() error = %v", err)
	}

	if session.interactionRespondCalls != 0 {
		t.Fatalf("interaction respond calls = %d, want 0", session.interactionRespondCalls)
	}
}

func TestHandleInteractionCreateRoutesComponent(t *testing.T) {
	session := &fakeSession{}
	bot := newBot(session)
	calls := 0
	bot.components.Handle(componentRoute{Namespace: "planning", Action: "join"}, func(*discordgo.InteractionCreate, componentRoute) error {
		calls++

		return nil
	})
	event := messageComponentInteraction("planning:join:session-123")

	err := bot.handleInteractionCreate(event)

	if err != nil {
		t.Fatalf("handleInteractionCreate() error = %v", err)
	}

	if calls != 1 {
		t.Fatalf("component handler calls = %d, want 1", calls)
	}
}

func TestHandleInteractionCreateRejectsInvalidComponentID(t *testing.T) {
	session := &fakeSession{}
	bot := newBot(session)
	event := messageComponentInteraction("invalid")

	err := bot.handleInteractionCreate(event)

	if !errors.Is(err, ErrComponentIDInvalid) {
		t.Fatalf("handleInteractionCreate() error = %v, want invalid component id", err)
	}
}

func TestHandleInteractionCreateReturnsResponseError(t *testing.T) {
	respondErr := errors.New("discord respond")
	session := &fakeSession{interactionRespondErr: respondErr}
	bot := newBot(session)

	err := bot.handleInteractionCreate(pingInteraction())

	if !errors.Is(err, respondErr) {
		t.Fatalf("handleInteractionCreate() error = %v, want response error", err)
	}
}

func TestHandleInteractionCreateIgnoresNilInteraction(t *testing.T) {
	session := &fakeSession{}
	bot := newBot(session)

	err := bot.handleInteractionCreate(nil)

	if err != nil {
		t.Fatalf("handleInteractionCreate() nil error = %v", err)
	}

	err = bot.handleInteractionCreate(&discordgo.InteractionCreate{})

	if err != nil {
		t.Fatalf("handleInteractionCreate() empty error = %v", err)
	}
}

func pingInteraction() *discordgo.InteractionCreate {
	return applicationCommandInteraction(pingCommandName)
}

func applicationCommandInteraction(name string) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:    "interaction-id",
			Token: "interaction-token",
			Type:  discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Name: name,
			},
		},
	}
}

func messageComponentInteraction(customID string) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:    "interaction-id",
			Token: "interaction-token",
			Type:  discordgo.InteractionMessageComponent,
			Data: discordgo.MessageComponentInteractionData{
				CustomID: customID,
			},
		},
	}
}
