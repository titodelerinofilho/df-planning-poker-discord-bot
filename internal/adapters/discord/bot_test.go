package discord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

func TestNewBotConfiguresMinimalIntents(t *testing.T) {
	session := &fakeSession{}

	newBot(session)

	if session.intents != discordgo.IntentsGuilds {
		t.Fatalf("intents = %v, want %v", session.intents, discordgo.IntentsGuilds)
	}
}

func TestNewBotKeepsHTTPClientTimeout(t *testing.T) {
	bot, err := NewBot("token")

	if err != nil {
		t.Fatalf("NewBot() error = %v", err)
	}

	session, ok := bot.session.(*discordSession)

	if !ok {
		t.Fatalf("session = %T, want *discordSession", bot.session)
	}

	if session.session.Client == nil || session.session.Client.Timeout <= 0 {
		t.Fatal("NewBot() did not keep an HTTP client timeout")
	}
}

func TestBotOpenOpensSessionOnce(t *testing.T) {
	session := &fakeSession{}
	bot := newBot(session)

	err := bot.Open(context.Background())

	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	err = bot.Open(context.Background())

	if err != nil {
		t.Fatalf("Open() second error = %v", err)
	}

	if session.openCalls != 1 {
		t.Fatalf("open calls = %d, want 1", session.openCalls)
	}
}

func TestBotOpenReturnsSessionError(t *testing.T) {
	openErr := errors.New("gateway down")
	session := &fakeSession{openErr: openErr}
	bot := newBot(session)

	err := bot.Open(context.Background())

	if !errors.Is(err, openErr) {
		t.Fatalf("Open() error = %v, want %v", err, openErr)
	}
}

func TestBotCloseClosesOpenedSessionOnce(t *testing.T) {
	session := &fakeSession{}
	bot := newBot(session)

	err := bot.Open(context.Background())

	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	err = bot.Close(context.Background())

	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	err = bot.Close(context.Background())

	if err != nil {
		t.Fatalf("Close() second error = %v", err)
	}

	if session.closeCalls != 1 {
		t.Fatalf("close calls = %d, want 1", session.closeCalls)
	}
}

func TestBotCloseReturnsSessionError(t *testing.T) {
	closeErr := errors.New("close gateway")
	session := &fakeSession{closeErr: closeErr}
	bot := newBot(session)

	err := bot.Open(context.Background())

	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	err = bot.Close(context.Background())

	if !errors.Is(err, closeErr) {
		t.Fatalf("Close() error = %v, want %v", err, closeErr)
	}
}

func TestBotOpenRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	session := &fakeSession{}
	bot := newBot(session)

	err := bot.Open(ctx)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Open() error = %v, want context.Canceled", err)
	}

	if session.openCalls != 0 {
		t.Fatalf("open calls = %d, want 0", session.openCalls)
	}
}

func TestNewBotRegistersInteractionHandler(t *testing.T) {
	session := &fakeSession{}

	newBot(session)

	if session.addInteractionCreateHandlerCalls != 1 {
		t.Fatalf("interaction handler calls = %d, want 1", session.addInteractionCreateHandlerCalls)
	}

	if session.interactionCreateHandler == nil {
		t.Fatal("interaction handler was not registered")
	}
}

type fakeSession struct {
	intents discordgo.Intent

	openCalls  int
	closeCalls int

	openErr  error
	closeErr error

	addInteractionCreateHandlerCalls int
	interactionCreateHandler         func(*discordgo.InteractionCreate)
	heartbeatLatency                 time.Duration
	interactionResponse              *discordgo.InteractionResponse
	interactionRespondCalls          int
	interactionRespondErr            error

	applicationID string
	guildID       string

	existingCommands []*discordgo.ApplicationCommand

	listCommandCalls   int
	createCommandCalls int
	editCommandCalls   int
	deleteCommandCalls int

	createdCommands  []*discordgo.ApplicationCommand
	editedCommands   []*discordgo.ApplicationCommand
	deletedCommandID []string

	listCommandsErr  error
	createCommandErr error
	editCommandErr   error
	deleteCommandErr error
}

func (session *fakeSession) Open() error {
	session.openCalls++

	return session.openErr
}

func (session *fakeSession) Close() error {
	session.closeCalls++

	return session.closeErr
}

func (session *fakeSession) SetIntents(intents discordgo.Intent) {
	session.intents = intents
}

func (session *fakeSession) AddInteractionCreateHandler(handler func(*discordgo.InteractionCreate)) func() {
	session.addInteractionCreateHandlerCalls++
	session.interactionCreateHandler = handler

	return func() {}
}

func (session *fakeSession) HeartbeatLatency() time.Duration {
	return session.heartbeatLatency
}

func (session *fakeSession) InteractionRespond(_ *discordgo.Interaction, response *discordgo.InteractionResponse, _ ...discordgo.RequestOption) error {
	session.interactionRespondCalls++
	session.interactionResponse = response

	return session.interactionRespondErr
}

func (session *fakeSession) ApplicationCommands(applicationID string, guildID string, _ ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error) {
	session.listCommandCalls++
	session.applicationID = applicationID
	session.guildID = guildID

	return session.existingCommands, session.listCommandsErr
}

func (session *fakeSession) ApplicationCommandCreate(applicationID string, guildID string, command *discordgo.ApplicationCommand, _ ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error) {
	session.createCommandCalls++
	session.applicationID = applicationID
	session.guildID = guildID
	session.createdCommands = append(session.createdCommands, command)

	return command, session.createCommandErr
}

func (session *fakeSession) ApplicationCommandEdit(applicationID string, guildID string, commandID string, command *discordgo.ApplicationCommand, _ ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error) {
	session.editCommandCalls++
	session.applicationID = applicationID
	session.guildID = guildID
	command.ID = commandID
	session.editedCommands = append(session.editedCommands, command)

	return command, session.editCommandErr
}

func (session *fakeSession) ApplicationCommandDelete(applicationID string, guildID string, commandID string, _ ...discordgo.RequestOption) error {
	session.deleteCommandCalls++
	session.applicationID = applicationID
	session.guildID = guildID
	session.deletedCommandID = append(session.deletedCommandID, commandID)

	return session.deleteCommandErr
}
