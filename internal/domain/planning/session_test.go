package planning

import (
	"errors"
	"testing"
	"time"
)

func TestNewSessionCreatesJoiningSession(t *testing.T) {
	input := validNewSessionInput()

	session, err := NewSession(input)

	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	if session.ID() != input.ID {
		t.Fatalf("ID() = %q, want %q", session.ID(), input.ID)
	}

	if session.GuildID() != input.GuildID {
		t.Fatalf("GuildID() = %q, want %q", session.GuildID(), input.GuildID)
	}

	if session.ChannelID() != input.ChannelID {
		t.Fatalf("ChannelID() = %q, want %q", session.ChannelID(), input.ChannelID)
	}

	if session.ThreadID() != input.ThreadID {
		t.Fatalf("ThreadID() = %q, want %q", session.ThreadID(), input.ThreadID)
	}

	if session.MessageID() != input.MessageID {
		t.Fatalf("MessageID() = %q, want %q", session.MessageID(), input.MessageID)
	}

	if session.Task() != input.Task {
		t.Fatalf("Task() = %#v, want %#v", session.Task(), input.Task)
	}

	if session.CreatorID() != input.CreatorID {
		t.Fatalf("CreatorID() = %q, want %q", session.CreatorID(), input.CreatorID)
	}

	if session.FacilitatorID() != input.FacilitatorID {
		t.Fatalf("FacilitatorID() = %q, want %q", session.FacilitatorID(), input.FacilitatorID)
	}

	if session.State() != SessionStateJoining {
		t.Fatalf("State() = %q, want JOINING", session.State())
	}

	if session.CurrentRoundNumber() != 0 {
		t.Fatalf("CurrentRoundNumber() = %d, want 0", session.CurrentRoundNumber())
	}

	if !session.CreatedAt().Equal(input.CreatedAt) {
		t.Fatalf("CreatedAt() = %s, want %s", session.CreatedAt(), input.CreatedAt)
	}

	if !session.ExpiresAt().Equal(input.ExpiresAt) {
		t.Fatalf("ExpiresAt() = %s, want %s", session.ExpiresAt(), input.ExpiresAt)
	}
}

func TestNewSessionTrimsTaskFields(t *testing.T) {
	input := validNewSessionInput()
	input.Task = Task{
		URL:   " https://example.com/issues/42 ",
		Title: " Criar sessao ",
	}

	session, err := NewSession(input)

	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	if session.Task().URL != "https://example.com/issues/42" {
		t.Fatalf("task url = %q, want trimmed url", session.Task().URL)
	}

	if session.Task().Title != "Criar sessao" {
		t.Fatalf("task title = %q, want trimmed title", session.Task().Title)
	}
}

func TestNewSessionKeepsScaleValues(t *testing.T) {
	input := validNewSessionInput()

	session, err := NewSession(input)

	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	values := session.Scale().Values()

	if len(values) != len(input.Scale.Values()) {
		t.Fatalf("scale values = %d, want %d", len(values), len(input.Scale.Values()))
	}

	if values[0].String() != "0" || values[len(values)-1].String() != SpecialEstimateCoffee {
		t.Fatalf("scale values = %#v, want modified fibonacci values", values)
	}
}

func TestNewSessionRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*NewSessionInput)
	}{
		{name: "missing id", mutate: func(input *NewSessionInput) { input.ID = "" }},
		{name: "missing guild", mutate: func(input *NewSessionInput) { input.GuildID = "" }},
		{name: "missing channel", mutate: func(input *NewSessionInput) { input.ChannelID = "" }},
		{name: "missing thread", mutate: func(input *NewSessionInput) { input.ThreadID = "" }},
		{name: "missing message", mutate: func(input *NewSessionInput) { input.MessageID = "" }},
		{name: "missing creator", mutate: func(input *NewSessionInput) { input.CreatorID = "" }},
		{name: "missing facilitator", mutate: func(input *NewSessionInput) { input.FacilitatorID = "" }},
		{name: "missing created at", mutate: func(input *NewSessionInput) { input.CreatedAt = time.Time{} }},
		{name: "missing expires at", mutate: func(input *NewSessionInput) { input.ExpiresAt = time.Time{} }},
		{name: "expires before created", mutate: func(input *NewSessionInput) { input.ExpiresAt = input.CreatedAt.Add(-time.Minute) }},
		{name: "missing task url", mutate: func(input *NewSessionInput) { input.Task.URL = "" }},
		{name: "invalid task url", mutate: func(input *NewSessionInput) { input.Task.URL = "not a url" }},
		{name: "missing task title", mutate: func(input *NewSessionInput) { input.Task.Title = "" }},
		{name: "missing scale", mutate: func(input *NewSessionInput) { input.Scale = Scale{} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validNewSessionInput()
			tt.mutate(&input)

			_, err := NewSession(input)

			if !errors.Is(err, ErrInvalidSession) {
				t.Fatalf("NewSession() error = %v, want ErrInvalidSession", err)
			}
		})
	}
}

func validNewSessionInput() NewSessionInput {
	createdAt := time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC)

	return NewSessionInput{
		ID:            SessionID("session-123"),
		GuildID:       GuildID("guild-123"),
		ChannelID:     ChannelID("channel-123"),
		ThreadID:      ThreadID("thread-123"),
		MessageID:     MessageID("message-123"),
		Task:          Task{URL: "https://example.com/issues/42", Title: "Criar sessao"},
		CreatorID:     DiscordUserID("creator-123"),
		FacilitatorID: DiscordUserID("facilitator-123"),
		Scale:         ModifiedFibonacciScale(),
		CreatedAt:     createdAt,
		ExpiresAt:     createdAt.Add(24 * time.Hour),
	}
}
