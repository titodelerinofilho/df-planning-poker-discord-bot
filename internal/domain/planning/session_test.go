package planning

import (
	"errors"
	"slices"
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

func TestSessionJoinParticipant(t *testing.T) {
	session := validSession(t)
	joinedAt := time.Date(2026, 8, 12, 11, 0, 0, 0, time.UTC)

	err := session.JoinParticipant(DiscordUserID("user-123"), " Ana Silva ", joinedAt)

	if err != nil {
		t.Fatalf("JoinParticipant() error = %v", err)
	}

	participants := session.Participants()

	if len(participants) != 1 {
		t.Fatalf("Participants() length = %d, want 1", len(participants))
	}

	participant := participants[0]

	if participant.SessionID() != session.ID() {
		t.Fatalf("participant session id = %q, want %q", participant.SessionID(), session.ID())
	}

	if participant.DiscordUserID() != DiscordUserID("user-123") {
		t.Fatalf("participant discord user id = %q, want user-123", participant.DiscordUserID())
	}

	if participant.DisplayName() != "Ana Silva" {
		t.Fatalf("participant display name = %q, want Ana Silva", participant.DisplayName())
	}

	if !participant.Active() {
		t.Fatal("participant active = false, want true")
	}

	if !participant.JoinedAt().Equal(joinedAt) {
		t.Fatalf("participant joined at = %s, want %s", participant.JoinedAt(), joinedAt)
	}

	if !participant.LeftAt().IsZero() {
		t.Fatalf("participant left at = %s, want zero time", participant.LeftAt())
	}
}

func TestSessionJoinParticipantRejectsDuplicateActiveParticipant(t *testing.T) {
	session := validSession(t)
	joinedAt := time.Date(2026, 8, 12, 11, 0, 0, 0, time.UTC)

	err := session.JoinParticipant(DiscordUserID("user-123"), "Ana Silva", joinedAt)

	if err != nil {
		t.Fatalf("JoinParticipant() error = %v", err)
	}

	err = session.JoinParticipant(DiscordUserID("user-123"), "Ana Silva", joinedAt.Add(time.Minute))

	if !errors.Is(err, ErrAlreadyParticipant) {
		t.Fatalf("JoinParticipant() error = %v, want ErrAlreadyParticipant", err)
	}
}

func TestSessionJoinParticipantRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		discordUserID DiscordUserID
		displayName   string
		joinedAt      time.Time
	}{
		{
			name:        "missing discord user id",
			displayName: "Ana Silva",
			joinedAt:    time.Date(2026, 8, 12, 11, 0, 0, 0, time.UTC),
		},
		{
			name:          "missing display name",
			discordUserID: DiscordUserID("user-123"),
			displayName:   " ",
			joinedAt:      time.Date(2026, 8, 12, 11, 0, 0, 0, time.UTC),
		},
		{
			name:          "missing joined at",
			discordUserID: DiscordUserID("user-123"),
			displayName:   "Ana Silva",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := validSession(t)

			err := session.JoinParticipant(tt.discordUserID, tt.displayName, tt.joinedAt)

			if !errors.Is(err, ErrInvalidParticipant) {
				t.Fatalf("JoinParticipant() error = %v, want ErrInvalidParticipant", err)
			}
		})
	}
}

func TestSessionJoinParticipantRejectsInvalidState(t *testing.T) {
	session := validSession(t)
	session.state = SessionStateVoting

	err := session.JoinParticipant(
		DiscordUserID("user-123"),
		"Ana Silva",
		time.Date(2026, 8, 12, 11, 0, 0, 0, time.UTC),
	)

	if !errors.Is(err, ErrSessionNotOpen) {
		t.Fatalf("JoinParticipant() error = %v, want ErrSessionNotOpen", err)
	}
}

func TestSessionLeaveParticipant(t *testing.T) {
	session := validSession(t)
	joinedAt := time.Date(2026, 8, 12, 11, 0, 0, 0, time.UTC)
	leftAt := joinedAt.Add(10 * time.Minute)

	err := session.JoinParticipant(DiscordUserID("user-123"), "Ana Silva", joinedAt)

	if err != nil {
		t.Fatalf("JoinParticipant() error = %v", err)
	}

	err = session.LeaveParticipant(DiscordUserID("user-123"), leftAt)

	if err != nil {
		t.Fatalf("LeaveParticipant() error = %v", err)
	}

	participants := session.Participants()

	if len(participants) != 1 {
		t.Fatalf("Participants() length = %d, want 1", len(participants))
	}

	participant := participants[0]

	if participant.Active() {
		t.Fatal("participant active = true, want false")
	}

	if !participant.LeftAt().Equal(leftAt) {
		t.Fatalf("participant left at = %s, want %s", participant.LeftAt(), leftAt)
	}

	if len(session.ActiveParticipants()) != 0 {
		t.Fatalf("ActiveParticipants() length = %d, want 0", len(session.ActiveParticipants()))
	}
}

func TestSessionLeaveParticipantAllowsRejoinWhileJoining(t *testing.T) {
	session := validSession(t)
	joinedAt := time.Date(2026, 8, 12, 11, 0, 0, 0, time.UTC)
	rejoinedAt := joinedAt.Add(20 * time.Minute)

	err := session.JoinParticipant(DiscordUserID("user-123"), "Ana Silva", joinedAt)

	if err != nil {
		t.Fatalf("JoinParticipant() error = %v", err)
	}

	err = session.LeaveParticipant(DiscordUserID("user-123"), joinedAt.Add(10*time.Minute))

	if err != nil {
		t.Fatalf("LeaveParticipant() error = %v", err)
	}

	err = session.JoinParticipant(DiscordUserID("user-123"), "Ana Silva", rejoinedAt)

	if err != nil {
		t.Fatalf("JoinParticipant() error = %v", err)
	}

	participants := session.ActiveParticipants()

	if len(participants) != 1 {
		t.Fatalf("ActiveParticipants() length = %d, want 1", len(participants))
	}

	if !participants[0].JoinedAt().Equal(rejoinedAt) {
		t.Fatalf("participant joined at = %s, want %s", participants[0].JoinedAt(), rejoinedAt)
	}

	if !participants[0].LeftAt().IsZero() {
		t.Fatalf("participant left at = %s, want zero time", participants[0].LeftAt())
	}
}

func TestSessionLeaveParticipantRejectsUnknownParticipant(t *testing.T) {
	session := validSession(t)

	err := session.LeaveParticipant(
		DiscordUserID("user-123"),
		time.Date(2026, 8, 12, 11, 0, 0, 0, time.UTC),
	)

	if !errors.Is(err, ErrParticipantNotFound) {
		t.Fatalf("LeaveParticipant() error = %v, want ErrParticipantNotFound", err)
	}
}

func TestSessionLeaveParticipantRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name          string
		discordUserID DiscordUserID
		leftAt        time.Time
	}{
		{
			name:   "missing discord user id",
			leftAt: time.Date(2026, 8, 12, 11, 0, 0, 0, time.UTC),
		},
		{
			name:          "missing left at",
			discordUserID: DiscordUserID("user-123"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := validSession(t)

			err := session.LeaveParticipant(tt.discordUserID, tt.leftAt)

			if !errors.Is(err, ErrInvalidParticipant) {
				t.Fatalf("LeaveParticipant() error = %v, want ErrInvalidParticipant", err)
			}
		})
	}
}

func TestSessionLeaveParticipantRejectsInvalidState(t *testing.T) {
	session := validSession(t)

	err := session.JoinParticipant(
		DiscordUserID("user-123"),
		"Ana Silva",
		time.Date(2026, 8, 12, 11, 0, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("JoinParticipant() error = %v", err)
	}

	session.state = SessionStateVoting

	err = session.LeaveParticipant(
		DiscordUserID("user-123"),
		time.Date(2026, 8, 12, 11, 10, 0, 0, time.UTC),
	)

	if !errors.Is(err, ErrSessionNotOpen) {
		t.Fatalf("LeaveParticipant() error = %v, want ErrSessionNotOpen", err)
	}
}

func TestSessionParticipantsAreOrderedByJoinTime(t *testing.T) {
	session := validSession(t)
	baseTime := time.Date(2026, 8, 12, 11, 0, 0, 0, time.UTC)

	err := session.JoinParticipant(DiscordUserID("user-3"), "Carol", baseTime.Add(2*time.Minute))

	if err != nil {
		t.Fatalf("JoinParticipant() error = %v", err)
	}

	err = session.JoinParticipant(DiscordUserID("user-1"), "Ana", baseTime)

	if err != nil {
		t.Fatalf("JoinParticipant() error = %v", err)
	}

	err = session.JoinParticipant(DiscordUserID("user-2"), "Bruno", baseTime.Add(time.Minute))

	if err != nil {
		t.Fatalf("JoinParticipant() error = %v", err)
	}

	participants := session.Participants()
	got := []DiscordUserID{
		participants[0].DiscordUserID(),
		participants[1].DiscordUserID(),
		participants[2].DiscordUserID(),
	}
	want := []DiscordUserID{
		DiscordUserID("user-1"),
		DiscordUserID("user-2"),
		DiscordUserID("user-3"),
	}

	if !slices.Equal(got, want) {
		t.Fatalf("Participants() order = %#v, want %#v", got, want)
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

func validSession(t *testing.T) Session {
	t.Helper()

	session, err := NewSession(validNewSessionInput())

	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	return session
}
