package discord

import (
	"context"
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestSyncGuildCommandsCreatesMissingCommands(t *testing.T) {
	session := &fakeSession{}
	bot := newBot(session)

	err := bot.SyncGuildCommands(context.Background(), "application-id", "guild-id", []CommandDefinition{
		{Name: "planning", Description: "Inicia uma sessao de planning poker"},
	})

	if err != nil {
		t.Fatalf("SyncGuildCommands() error = %v", err)
	}

	if session.listCommandCalls != 1 {
		t.Fatalf("list calls = %d, want 1", session.listCommandCalls)
	}

	if session.createCommandCalls != 1 {
		t.Fatalf("create calls = %d, want 1", session.createCommandCalls)
	}

	created := session.createdCommands[0]

	if created.Type != discordgo.ChatApplicationCommand {
		t.Fatalf("created type = %v, want chat command", created.Type)
	}

	if created.Name != "planning" {
		t.Fatalf("created name = %q, want planning", created.Name)
	}

	if created.Description != "Inicia uma sessao de planning poker" {
		t.Fatalf("created description = %q", created.Description)
	}
}

func TestSyncGuildCommandsUpdatesChangedCommands(t *testing.T) {
	session := &fakeSession{
		existingCommands: []*discordgo.ApplicationCommand{
			{
				ID:          "command-id",
				Type:        discordgo.ChatApplicationCommand,
				Name:        "planning",
				Description: "Descricao antiga",
			},
		},
	}
	bot := newBot(session)

	err := bot.SyncGuildCommands(context.Background(), "application-id", "guild-id", []CommandDefinition{
		{Name: "planning", Description: "Descricao nova"},
	})

	if err != nil {
		t.Fatalf("SyncGuildCommands() error = %v", err)
	}

	if session.editCommandCalls != 1 {
		t.Fatalf("edit calls = %d, want 1", session.editCommandCalls)
	}

	edited := session.editedCommands[0]

	if edited.ID != "command-id" {
		t.Fatalf("edited id = %q, want command-id", edited.ID)
	}

	if edited.Description != "Descricao nova" {
		t.Fatalf("edited description = %q, want Descricao nova", edited.Description)
	}
}

func TestSyncGuildCommandsKeepsMatchingCommands(t *testing.T) {
	session := &fakeSession{
		existingCommands: []*discordgo.ApplicationCommand{
			{
				ID:          "command-id",
				Type:        discordgo.ChatApplicationCommand,
				Name:        "planning",
				Description: "Inicia uma sessao de planning poker",
			},
		},
	}
	bot := newBot(session)

	err := bot.SyncGuildCommands(context.Background(), "application-id", "guild-id", []CommandDefinition{
		{Name: "planning", Description: "Inicia uma sessao de planning poker"},
	})

	if err != nil {
		t.Fatalf("SyncGuildCommands() error = %v", err)
	}

	if session.createCommandCalls != 0 {
		t.Fatalf("create calls = %d, want 0", session.createCommandCalls)
	}

	if session.editCommandCalls != 0 {
		t.Fatalf("edit calls = %d, want 0", session.editCommandCalls)
	}
}

func TestSyncGuildCommandsDeletesOnlyManagedCommands(t *testing.T) {
	session := &fakeSession{
		existingCommands: []*discordgo.ApplicationCommand{
			{ID: "old-id", Type: discordgo.ChatApplicationCommand, Name: "planning", Description: "old"},
			{ID: "manual-id", Type: discordgo.ChatApplicationCommand, Name: "manual", Description: "manual"},
		},
	}
	bot := newBot(session)

	err := bot.syncGuildCommands(
		context.Background(),
		"application-id",
		"guild-id",
		nil,
		[]string{"planning"},
	)

	if err != nil {
		t.Fatalf("syncGuildCommands() error = %v", err)
	}

	if session.deleteCommandCalls != 1 {
		t.Fatalf("delete calls = %d, want 1", session.deleteCommandCalls)
	}

	if session.deletedCommandID[0] != "old-id" {
		t.Fatalf("deleted id = %q, want old-id", session.deletedCommandID[0])
	}
}

func TestSyncGuildCommandsReturnsListError(t *testing.T) {
	listErr := errors.New("discord list")
	session := &fakeSession{listCommandsErr: listErr}
	bot := newBot(session)

	err := bot.SyncGuildCommands(context.Background(), "application-id", "guild-id", nil)

	if !errors.Is(err, listErr) {
		t.Fatalf("SyncGuildCommands() error = %v, want list error", err)
	}
}

func TestSyncGuildCommandsRejectsDuplicateCommandName(t *testing.T) {
	session := &fakeSession{}
	bot := newBot(session)

	err := bot.SyncGuildCommands(context.Background(), "application-id", "guild-id", []CommandDefinition{
		{Name: "planning", Description: "Primeira"},
		{Name: " planning ", Description: "Segunda"},
	})

	if err == nil {
		t.Fatal("SyncGuildCommands() error = nil, want duplicate command error")
	}

	if session.listCommandCalls != 0 {
		t.Fatalf("list calls = %d, want 0", session.listCommandCalls)
	}
}

func TestSyncGuildCommandsRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	session := &fakeSession{}
	bot := newBot(session)

	err := bot.SyncGuildCommands(ctx, "application-id", "guild-id", nil)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SyncGuildCommands() error = %v, want context.Canceled", err)
	}

	if session.listCommandCalls != 0 {
		t.Fatalf("list calls = %d, want 0", session.listCommandCalls)
	}
}
