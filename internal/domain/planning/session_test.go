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

	if _, ok := session.CurrentRound(); ok {
		t.Fatal("CurrentRound() ok = true, want false")
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

func TestSessionCloseParticipantsOpensVotingRound(t *testing.T) {
	session := validSession(t)
	openedAt := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)

	joinParticipant(t, &session, DiscordUserID("user-1"), "Ana", openedAt.Add(-2*time.Minute))
	joinParticipant(t, &session, DiscordUserID("user-2"), "Bruno", openedAt.Add(-time.Minute))

	err := session.CloseParticipants(2, RoundID("round-1"), openedAt)

	if err != nil {
		t.Fatalf("CloseParticipants() error = %v", err)
	}

	if session.State() != SessionStateVoting {
		t.Fatalf("State() = %q, want VOTING", session.State())
	}

	if session.CurrentRoundNumber() != 1 {
		t.Fatalf("CurrentRoundNumber() = %d, want 1", session.CurrentRoundNumber())
	}

	round, ok := session.CurrentRound()

	if !ok {
		t.Fatal("CurrentRound() ok = false, want true")
	}

	if round.ID() != RoundID("round-1") {
		t.Fatalf("round ID() = %q, want round-1", round.ID())
	}

	if round.SessionID() != session.ID() {
		t.Fatalf("round SessionID() = %q, want %q", round.SessionID(), session.ID())
	}

	if round.Number() != 1 {
		t.Fatalf("round Number() = %d, want 1", round.Number())
	}

	if round.State() != RoundStateOpen {
		t.Fatalf("round State() = %q, want OPEN", round.State())
	}

	if !round.OpenedAt().Equal(openedAt) {
		t.Fatalf("round OpenedAt() = %s, want %s", round.OpenedAt(), openedAt)
	}
}

func TestSessionCloseParticipantsRejectsNotEnoughParticipants(t *testing.T) {
	session := validSession(t)
	openedAt := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)

	joinParticipant(t, &session, DiscordUserID("user-1"), "Ana", openedAt.Add(-time.Minute))

	err := session.CloseParticipants(2, RoundID("round-1"), openedAt)

	if !errors.Is(err, ErrNotEnoughParticipants) {
		t.Fatalf("CloseParticipants() error = %v, want ErrNotEnoughParticipants", err)
	}

	if session.State() != SessionStateJoining {
		t.Fatalf("State() = %q, want JOINING", session.State())
	}

	if _, ok := session.CurrentRound(); ok {
		t.Fatal("CurrentRound() ok = true, want false")
	}
}

func TestSessionCloseParticipantsIgnoresInactiveParticipants(t *testing.T) {
	session := validSession(t)
	openedAt := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)

	joinParticipant(t, &session, DiscordUserID("user-1"), "Ana", openedAt.Add(-2*time.Minute))

	err := session.LeaveParticipant(DiscordUserID("user-1"), openedAt.Add(-time.Minute))

	if err != nil {
		t.Fatalf("LeaveParticipant() error = %v", err)
	}

	err = session.CloseParticipants(1, RoundID("round-1"), openedAt)

	if !errors.Is(err, ErrNotEnoughParticipants) {
		t.Fatalf("CloseParticipants() error = %v, want ErrNotEnoughParticipants", err)
	}
}

func TestSessionCloseParticipantsRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name                string
		minimumParticipants int
		roundID             RoundID
		openedAt            time.Time
		want                error
	}{
		{
			name:                "invalid minimum participants",
			minimumParticipants: 0,
			roundID:             RoundID("round-1"),
			openedAt:            time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC),
			want:                ErrInvalidParticipant,
		},
		{
			name:                "missing round id",
			minimumParticipants: 1,
			openedAt:            time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC),
			want:                ErrInvalidRound,
		},
		{
			name:                "missing opened at",
			minimumParticipants: 1,
			roundID:             RoundID("round-1"),
			want:                ErrInvalidRound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := validSession(t)
			joinParticipant(
				t,
				&session,
				DiscordUserID("user-1"),
				"Ana",
				time.Date(2026, 8, 12, 11, 59, 0, 0, time.UTC),
			)

			err := session.CloseParticipants(tt.minimumParticipants, tt.roundID, tt.openedAt)

			if !errors.Is(err, tt.want) {
				t.Fatalf("CloseParticipants() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestSessionCloseParticipantsRejectsInvalidState(t *testing.T) {
	session := validSession(t)
	openedAt := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)

	joinParticipant(t, &session, DiscordUserID("user-1"), "Ana", openedAt.Add(-time.Minute))

	err := session.CloseParticipants(1, RoundID("round-1"), openedAt)

	if err != nil {
		t.Fatalf("CloseParticipants() error = %v", err)
	}

	err = session.CloseParticipants(1, RoundID("round-2"), openedAt.Add(time.Minute))

	if !errors.Is(err, ErrSessionNotOpen) {
		t.Fatalf("CloseParticipants() error = %v, want ErrSessionNotOpen", err)
	}
}

func TestSessionCastVoteRecordsVote(t *testing.T) {
	session := votingSession(t, 2)
	castAt := time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC)

	err := session.CastVote(DiscordUserID("user-1"), NewEstimate("5"), castAt)

	if err != nil {
		t.Fatalf("CastVote() error = %v", err)
	}

	if session.State() != SessionStateVoting {
		t.Fatalf("State() = %q, want VOTING", session.State())
	}

	round, ok := session.CurrentRound()

	if !ok {
		t.Fatal("CurrentRound() ok = false, want true")
	}

	if round.State() != RoundStateOpen {
		t.Fatalf("round State() = %q, want OPEN", round.State())
	}

	if round.VoteCount() != 1 {
		t.Fatalf("round VoteCount() = %d, want 1", round.VoteCount())
	}

	if !round.HasVoteFrom(DiscordUserID("user-1")) {
		t.Fatal("round HasVoteFrom(user-1) = false, want true")
	}

	vote := round.votes[DiscordUserID("user-1")]

	if vote.RoundID() != round.ID() {
		t.Fatalf("vote RoundID() = %q, want %q", vote.RoundID(), round.ID())
	}

	if vote.Estimate().String() != "5" {
		t.Fatalf("vote Estimate() = %q, want 5", vote.Estimate())
	}

	if !vote.FirstCastAt().Equal(castAt) {
		t.Fatalf("vote FirstCastAt() = %s, want %s", vote.FirstCastAt(), castAt)
	}

	if !vote.LastCastAt().Equal(castAt) {
		t.Fatalf("vote LastCastAt() = %s, want %s", vote.LastCastAt(), castAt)
	}
}

func TestSessionCastVoteMarksRoundReadyWhenAllParticipantsVoted(t *testing.T) {
	session := votingSession(t, 2)
	baseTime := time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC)

	err := session.CastVote(DiscordUserID("user-1"), NewEstimate("5"), baseTime)

	if err != nil {
		t.Fatalf("CastVote() error = %v", err)
	}

	err = session.CastVote(DiscordUserID("user-2"), NewEstimate("8"), baseTime.Add(time.Minute))

	if err != nil {
		t.Fatalf("CastVote() error = %v", err)
	}

	if session.State() != SessionStateReadyToReveal {
		t.Fatalf("State() = %q, want READY_TO_REVEAL", session.State())
	}

	round, ok := session.CurrentRound()

	if !ok {
		t.Fatal("CurrentRound() ok = false, want true")
	}

	if round.State() != RoundStateReady {
		t.Fatalf("round State() = %q, want READY", round.State())
	}

	got := round.VotedParticipantIDs()
	want := []DiscordUserID{
		DiscordUserID("user-1"),
		DiscordUserID("user-2"),
	}

	if !slices.Equal(got, want) {
		t.Fatalf("round VotedParticipantIDs() = %#v, want %#v", got, want)
	}
}

func TestSessionCastVoteAllowsChangeBeforeReveal(t *testing.T) {
	session := votingSession(t, 1)
	firstCastAt := time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC)
	secondCastAt := firstCastAt.Add(2 * time.Minute)

	err := session.CastVote(DiscordUserID("user-1"), NewEstimate("5"), firstCastAt)

	if err != nil {
		t.Fatalf("CastVote() error = %v", err)
	}

	err = session.CastVote(DiscordUserID("user-1"), NewEstimate("8"), secondCastAt)

	if err != nil {
		t.Fatalf("CastVote() error = %v", err)
	}

	round, ok := session.CurrentRound()

	if !ok {
		t.Fatal("CurrentRound() ok = false, want true")
	}

	if round.VoteCount() != 1 {
		t.Fatalf("round VoteCount() = %d, want 1", round.VoteCount())
	}

	vote := round.votes[DiscordUserID("user-1")]

	if vote.Estimate().String() != "8" {
		t.Fatalf("vote Estimate() = %q, want 8", vote.Estimate())
	}

	if !vote.FirstCastAt().Equal(firstCastAt) {
		t.Fatalf("vote FirstCastAt() = %s, want %s", vote.FirstCastAt(), firstCastAt)
	}

	if !vote.LastCastAt().Equal(secondCastAt) {
		t.Fatalf("vote LastCastAt() = %s, want %s", vote.LastCastAt(), secondCastAt)
	}

	if session.State() != SessionStateReadyToReveal {
		t.Fatalf("State() = %q, want READY_TO_REVEAL", session.State())
	}
}

func TestSessionCastVoteRejectsUnknownParticipant(t *testing.T) {
	session := votingSession(t, 1)

	err := session.CastVote(
		DiscordUserID("unknown-user"),
		NewEstimate("5"),
		time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC),
	)

	if !errors.Is(err, ErrParticipantNotFound) {
		t.Fatalf("CastVote() error = %v, want ErrParticipantNotFound", err)
	}
}

func TestSessionCastVoteRejectsInactiveParticipant(t *testing.T) {
	session := validSession(t)
	openedAt := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)

	joinParticipant(t, &session, DiscordUserID("user-1"), "Ana", openedAt.Add(-2*time.Minute))
	joinParticipant(t, &session, DiscordUserID("user-2"), "Bruno", openedAt.Add(-time.Minute))

	err := session.LeaveParticipant(DiscordUserID("user-1"), openedAt.Add(-30*time.Second))

	if err != nil {
		t.Fatalf("LeaveParticipant() error = %v", err)
	}

	err = session.CloseParticipants(1, RoundID("round-1"), openedAt)

	if err != nil {
		t.Fatalf("CloseParticipants() error = %v", err)
	}

	err = session.CastVote(DiscordUserID("user-1"), NewEstimate("5"), openedAt.Add(10*time.Minute))

	if !errors.Is(err, ErrParticipantNotFound) {
		t.Fatalf("CastVote() error = %v, want ErrParticipantNotFound", err)
	}
}

func TestSessionCastVoteRejectsInvalidEstimate(t *testing.T) {
	session := votingSession(t, 1)

	err := session.CastVote(
		DiscordUserID("user-1"),
		NewEstimate("100"),
		time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC),
	)

	if !errors.Is(err, ErrInvalidEstimate) {
		t.Fatalf("CastVote() error = %v, want ErrInvalidEstimate", err)
	}
}

func TestSessionCastVoteRejectsInvalidState(t *testing.T) {
	session := validSession(t)

	err := session.CastVote(
		DiscordUserID("user-1"),
		NewEstimate("5"),
		time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC),
	)

	if !errors.Is(err, ErrVoteNotAllowed) {
		t.Fatalf("CastVote() error = %v, want ErrVoteNotAllowed", err)
	}
}

func TestSessionCastVoteRejectsRevealedRound(t *testing.T) {
	session := votingSession(t, 1)
	session.state = SessionStateRevealed
	session.currentRound.state = RoundStateRevealed

	err := session.CastVote(
		DiscordUserID("user-1"),
		NewEstimate("5"),
		time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC),
	)

	if !errors.Is(err, ErrVoteNotAllowed) {
		t.Fatalf("CastVote() error = %v, want ErrVoteNotAllowed", err)
	}
}

func TestSessionCastVoteRejectsMissingCurrentRound(t *testing.T) {
	session := validSession(t)
	session.state = SessionStateVoting

	err := session.CastVote(
		DiscordUserID("user-1"),
		NewEstimate("5"),
		time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC),
	)

	if !errors.Is(err, ErrInvalidRound) {
		t.Fatalf("CastVote() error = %v, want ErrInvalidRound", err)
	}
}

func TestSessionCastVoteRejectsMissingCastAt(t *testing.T) {
	session := votingSession(t, 1)

	err := session.CastVote(DiscordUserID("user-1"), NewEstimate("5"), time.Time{})

	if !errors.Is(err, ErrVoteNotAllowed) {
		t.Fatalf("CastVote() error = %v, want ErrVoteNotAllowed", err)
	}
}

func TestSessionRevealRoundReturnsStructuredResult(t *testing.T) {
	session := votingSession(t, 3)
	baseTime := time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC)
	revealedAt := baseTime.Add(5 * time.Minute)

	castVote(t, &session, DiscordUserID("user-1"), "5", baseTime)
	castVote(t, &session, DiscordUserID("user-2"), "8", baseTime.Add(time.Minute))
	castVote(t, &session, DiscordUserID("user-3"), SpecialEstimateUnknown, baseTime.Add(2*time.Minute))

	result, err := session.RevealRound(revealedAt)

	if err != nil {
		t.Fatalf("RevealRound() error = %v", err)
	}

	if session.State() != SessionStateRevealed {
		t.Fatalf("State() = %q, want REVEALED", session.State())
	}

	round, ok := session.CurrentRound()

	if !ok {
		t.Fatal("CurrentRound() ok = false, want true")
	}

	if round.State() != RoundStateRevealed {
		t.Fatalf("round State() = %q, want REVEALED", round.State())
	}

	if !round.RevealedAt().Equal(revealedAt) {
		t.Fatalf("round RevealedAt() = %s, want %s", round.RevealedAt(), revealedAt)
	}

	roundStatistics, ok := round.NumericStatistics()

	if !ok {
		t.Fatal("round NumericStatistics() ok = false, want true")
	}

	if !roundStatistics.HasNumericResult {
		t.Fatal("round statistics HasNumericResult = false, want true")
	}

	if result.RoundID() != round.ID() {
		t.Fatalf("result RoundID() = %q, want %q", result.RoundID(), round.ID())
	}

	if result.RoundNumber() != round.Number() {
		t.Fatalf("result RoundNumber() = %d, want %d", result.RoundNumber(), round.Number())
	}

	if !result.RevealedAt().Equal(revealedAt) {
		t.Fatalf("result RevealedAt() = %s, want %s", result.RevealedAt(), revealedAt)
	}

	votes := result.Votes()

	if len(votes) != 3 {
		t.Fatalf("result Votes() length = %d, want 3", len(votes))
	}

	assertRevealedVote(t, votes[0], DiscordUserID("user-1"), "User 1", "5")
	assertRevealedVote(t, votes[1], DiscordUserID("user-2"), "User 2", "8")
	assertRevealedVote(t, votes[2], DiscordUserID("user-3"), "User 3", SpecialEstimateUnknown)

	statistics := result.NumericStatistics()

	if !statistics.HasNumericResult {
		t.Fatal("result statistics HasNumericResult = false, want true")
	}

	if statistics.Min != 5 {
		t.Fatalf("result statistics Min = %d, want 5", statistics.Min)
	}

	if statistics.Max != 8 {
		t.Fatalf("result statistics Max = %d, want 8", statistics.Max)
	}

	if statistics.Median != 6.5 {
		t.Fatalf("result statistics Median = %f, want 6.5", statistics.Median)
	}

	if !slices.Equal(statistics.Mode, []int{5, 8}) {
		t.Fatalf("result statistics Mode = %#v, want [5 8]", statistics.Mode)
	}
}

func TestSessionRevealRoundReturnsVotesCopy(t *testing.T) {
	session := votingSession(t, 1)
	castVote(t, &session, DiscordUserID("user-1"), "5", time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC))

	result, err := session.RevealRound(time.Date(2026, 8, 12, 12, 15, 0, 0, time.UTC))

	if err != nil {
		t.Fatalf("RevealRound() error = %v", err)
	}

	votes := result.Votes()
	votes[0] = RevealedVote{discordUserID: DiscordUserID("changed")}

	if result.Votes()[0].DiscordUserID() != DiscordUserID("user-1") {
		t.Fatalf("result Votes() leaked mutable slice")
	}
}

func TestSessionRevealRoundHandlesNoNumericResult(t *testing.T) {
	session := votingSession(t, 2)
	baseTime := time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC)

	castVote(t, &session, DiscordUserID("user-1"), SpecialEstimateUnknown, baseTime)
	castVote(t, &session, DiscordUserID("user-2"), SpecialEstimateCoffee, baseTime.Add(time.Minute))

	result, err := session.RevealRound(baseTime.Add(5 * time.Minute))

	if err != nil {
		t.Fatalf("RevealRound() error = %v", err)
	}

	if result.NumericStatistics().HasNumericResult {
		t.Fatal("result statistics HasNumericResult = true, want false")
	}
}

func TestSessionRevealRoundRejectsNotReadySession(t *testing.T) {
	session := votingSession(t, 2)
	castVote(t, &session, DiscordUserID("user-1"), "5", time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC))

	_, err := session.RevealRound(time.Date(2026, 8, 12, 12, 15, 0, 0, time.UTC))

	if !errors.Is(err, ErrRevealNotAllowed) {
		t.Fatalf("RevealRound() error = %v, want ErrRevealNotAllowed", err)
	}

	if session.State() != SessionStateVoting {
		t.Fatalf("State() = %q, want VOTING", session.State())
	}
}

func TestSessionRevealRoundRejectsMissingCurrentRound(t *testing.T) {
	session := validSession(t)
	session.state = SessionStateReadyToReveal

	_, err := session.RevealRound(time.Date(2026, 8, 12, 12, 15, 0, 0, time.UTC))

	if !errors.Is(err, ErrInvalidRound) {
		t.Fatalf("RevealRound() error = %v, want ErrInvalidRound", err)
	}
}

func TestSessionRevealRoundRejectsRoundNotReady(t *testing.T) {
	session := votingSession(t, 1)
	castVote(t, &session, DiscordUserID("user-1"), "5", time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC))
	session.currentRound.state = RoundStateOpen

	_, err := session.RevealRound(time.Date(2026, 8, 12, 12, 15, 0, 0, time.UTC))

	if !errors.Is(err, ErrRevealNotAllowed) {
		t.Fatalf("RevealRound() error = %v, want ErrRevealNotAllowed", err)
	}
}

func TestSessionRevealRoundRejectsMissingRevealedAt(t *testing.T) {
	session := votingSession(t, 1)
	castVote(t, &session, DiscordUserID("user-1"), "5", time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC))

	_, err := session.RevealRound(time.Time{})

	if !errors.Is(err, ErrRevealNotAllowed) {
		t.Fatalf("RevealRound() error = %v, want ErrRevealNotAllowed", err)
	}
}

func TestSessionRevealRoundFreezesVotes(t *testing.T) {
	session := votingSession(t, 1)
	castVote(t, &session, DiscordUserID("user-1"), "5", time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC))

	_, err := session.RevealRound(time.Date(2026, 8, 12, 12, 15, 0, 0, time.UTC))

	if err != nil {
		t.Fatalf("RevealRound() error = %v", err)
	}

	err = session.CastVote(DiscordUserID("user-1"), NewEstimate("8"), time.Date(2026, 8, 12, 12, 16, 0, 0, time.UTC))

	if !errors.Is(err, ErrVoteNotAllowed) {
		t.Fatalf("CastVote() error = %v, want ErrVoteNotAllowed", err)
	}
}

func TestSessionRestartRoundClosesPreviousRoundAndOpensNext(t *testing.T) {
	session := revealedSession(t, 2)
	openedAt := time.Date(2026, 8, 12, 12, 20, 0, 0, time.UTC)

	err := session.RestartRound(RoundID("round-2"), openedAt)

	if err != nil {
		t.Fatalf("RestartRound() error = %v", err)
	}

	if session.State() != SessionStateVoting {
		t.Fatalf("State() = %q, want VOTING", session.State())
	}

	if session.CurrentRoundNumber() != 2 {
		t.Fatalf("CurrentRoundNumber() = %d, want 2", session.CurrentRoundNumber())
	}

	round, ok := session.CurrentRound()

	if !ok {
		t.Fatal("CurrentRound() ok = false, want true")
	}

	if round.ID() != RoundID("round-2") {
		t.Fatalf("round ID() = %q, want round-2", round.ID())
	}

	if round.Number() != 2 {
		t.Fatalf("round Number() = %d, want 2", round.Number())
	}

	if round.State() != RoundStateOpen {
		t.Fatalf("round State() = %q, want OPEN", round.State())
	}

	if !round.OpenedAt().Equal(openedAt) {
		t.Fatalf("round OpenedAt() = %s, want %s", round.OpenedAt(), openedAt)
	}

	if round.VoteCount() != 0 {
		t.Fatalf("round VoteCount() = %d, want 0", round.VoteCount())
	}

	rounds := session.Rounds()

	if len(rounds) != 2 {
		t.Fatalf("Rounds() length = %d, want 2", len(rounds))
	}

	previousRound := rounds[0]

	if previousRound.ID() != RoundID("round-1") {
		t.Fatalf("previous round ID() = %q, want round-1", previousRound.ID())
	}

	if previousRound.State() != RoundStateClosed {
		t.Fatalf("previous round State() = %q, want CLOSED", previousRound.State())
	}

	if !previousRound.ClosedAt().Equal(openedAt) {
		t.Fatalf("previous round ClosedAt() = %s, want %s", previousRound.ClosedAt(), openedAt)
	}

	statistics, ok := previousRound.NumericStatistics()

	if !ok {
		t.Fatal("previous round NumericStatistics() ok = false, want true")
	}

	if !statistics.HasNumericResult {
		t.Fatal("previous round statistics HasNumericResult = false, want true")
	}

	if len(session.ActiveParticipants()) != 2 {
		t.Fatalf("ActiveParticipants() length = %d, want 2", len(session.ActiveParticipants()))
	}
}

func TestSessionRestartRoundAllowsVotingInNewRound(t *testing.T) {
	session := revealedSession(t, 1)

	err := session.RestartRound(
		RoundID("round-2"),
		time.Date(2026, 8, 12, 12, 20, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("RestartRound() error = %v", err)
	}

	err = session.CastVote(
		DiscordUserID("user-1"),
		NewEstimate("8"),
		time.Date(2026, 8, 12, 12, 21, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("CastVote() error = %v", err)
	}

	if session.State() != SessionStateReadyToReveal {
		t.Fatalf("State() = %q, want READY_TO_REVEAL", session.State())
	}

	round, ok := session.CurrentRound()

	if !ok {
		t.Fatal("CurrentRound() ok = false, want true")
	}

	if !round.HasVoteFrom(DiscordUserID("user-1")) {
		t.Fatal("round HasVoteFrom(user-1) = false, want true")
	}
}

func TestSessionRestartRoundReturnsRoundsCopy(t *testing.T) {
	session := revealedSession(t, 1)

	err := session.RestartRound(
		RoundID("round-2"),
		time.Date(2026, 8, 12, 12, 20, 0, 0, time.UTC),
	)

	if err != nil {
		t.Fatalf("RestartRound() error = %v", err)
	}

	rounds := session.Rounds()
	rounds[0] = Round{id: RoundID("changed")}

	if session.Rounds()[0].ID() != RoundID("round-1") {
		t.Fatal("Rounds() leaked mutable slice")
	}
}

func TestSessionRestartRoundRejectsNotRevealedSession(t *testing.T) {
	session := votingSession(t, 1)

	err := session.RestartRound(
		RoundID("round-2"),
		time.Date(2026, 8, 12, 12, 20, 0, 0, time.UTC),
	)

	if !errors.Is(err, ErrRestartRoundNotAllowed) {
		t.Fatalf("RestartRound() error = %v, want ErrRestartRoundNotAllowed", err)
	}
}

func TestSessionRestartRoundRejectsMissingCurrentRound(t *testing.T) {
	session := validSession(t)
	session.state = SessionStateRevealed

	err := session.RestartRound(
		RoundID("round-2"),
		time.Date(2026, 8, 12, 12, 20, 0, 0, time.UTC),
	)

	if !errors.Is(err, ErrInvalidRound) {
		t.Fatalf("RestartRound() error = %v, want ErrInvalidRound", err)
	}
}

func TestSessionRestartRoundRejectsCurrentRoundNotRevealed(t *testing.T) {
	session := revealedSession(t, 1)
	session.currentRound.state = RoundStateOpen

	err := session.RestartRound(
		RoundID("round-2"),
		time.Date(2026, 8, 12, 12, 20, 0, 0, time.UTC),
	)

	if !errors.Is(err, ErrRestartRoundNotAllowed) {
		t.Fatalf("RestartRound() error = %v, want ErrRestartRoundNotAllowed", err)
	}
}

func TestSessionRestartRoundRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		roundID  RoundID
		openedAt time.Time
	}{
		{
			name:     "missing round id",
			openedAt: time.Date(2026, 8, 12, 12, 20, 0, 0, time.UTC),
		},
		{
			name:    "missing opened at",
			roundID: RoundID("round-2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := revealedSession(t, 1)

			err := session.RestartRound(tt.roundID, tt.openedAt)

			if !errors.Is(err, ErrInvalidRound) {
				t.Fatalf("RestartRound() error = %v, want ErrInvalidRound", err)
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

func validSession(t *testing.T) Session {
	t.Helper()

	session, err := NewSession(validNewSessionInput())

	if err != nil {
		t.Fatalf("NewSession() error = %v", err)
	}

	return session
}

func joinParticipant(t *testing.T, session *Session, discordUserID DiscordUserID, displayName string, joinedAt time.Time) {
	t.Helper()

	err := session.JoinParticipant(discordUserID, displayName, joinedAt)

	if err != nil {
		t.Fatalf("JoinParticipant() error = %v", err)
	}
}

func castVote(t *testing.T, session *Session, discordUserID DiscordUserID, estimate string, castAt time.Time) {
	t.Helper()

	err := session.CastVote(discordUserID, NewEstimate(estimate), castAt)

	if err != nil {
		t.Fatalf("CastVote() error = %v", err)
	}
}

func assertRevealedVote(
	t *testing.T,
	vote RevealedVote,
	discordUserID DiscordUserID,
	displayName string,
	estimate string,
) {
	t.Helper()

	if vote.DiscordUserID() != discordUserID {
		t.Fatalf("vote DiscordUserID() = %q, want %q", vote.DiscordUserID(), discordUserID)
	}

	if vote.DisplayName() != displayName {
		t.Fatalf("vote DisplayName() = %q, want %q", vote.DisplayName(), displayName)
	}

	if vote.Estimate().String() != estimate {
		t.Fatalf("vote Estimate() = %q, want %q", vote.Estimate(), estimate)
	}
}

func votingSession(t *testing.T, participantCount int) Session {
	t.Helper()

	session := validSession(t)
	openedAt := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)

	for index := range participantCount {
		discordUserID := DiscordUserID("user-" + string(rune('1'+index)))
		displayName := "User " + string(rune('1'+index))

		joinedAt := openedAt.Add(-time.Duration(participantCount-index) * time.Minute)

		joinParticipant(t, &session, discordUserID, displayName, joinedAt)
	}

	err := session.CloseParticipants(participantCount, RoundID("round-1"), openedAt)

	if err != nil {
		t.Fatalf("CloseParticipants() error = %v", err)
	}

	return session
}

func revealedSession(t *testing.T, participantCount int) Session {
	t.Helper()

	session := votingSession(t, participantCount)
	baseTime := time.Date(2026, 8, 12, 12, 10, 0, 0, time.UTC)

	for index := range participantCount {
		discordUserID := DiscordUserID("user-" + string(rune('1'+index)))

		castVote(t, &session, discordUserID, "5", baseTime.Add(time.Duration(index)*time.Minute))
	}

	_, err := session.RevealRound(baseTime.Add(5 * time.Minute))

	if err != nil {
		t.Fatalf("RevealRound() error = %v", err)
	}

	return session
}
