package planning

import (
	"slices"
	"strings"
	"time"
)

type RoundID string

type RoundState string

const (
	RoundStateOpen     RoundState = "OPEN"
	RoundStateReady    RoundState = "READY"
	RoundStateRevealed RoundState = "REVEALED"
	RoundStateClosed   RoundState = "CLOSED"
)

type Round struct {
	id                   RoundID
	sessionID            SessionID
	number               int
	state                RoundState
	votes                map[DiscordUserID]Vote
	numericStatistics    NumericStatistics
	hasNumericStatistics bool
	openedAt             time.Time
	revealedAt           time.Time
	closedAt             time.Time
}

func newFirstRound(id RoundID, sessionID SessionID, openedAt time.Time) Round {
	return newRound(id, sessionID, 1, openedAt)
}

func newRound(id RoundID, sessionID SessionID, number int, openedAt time.Time) Round {
	return Round{
		id:        id,
		sessionID: sessionID,
		number:    number,
		state:     RoundStateOpen,
		votes:     make(map[DiscordUserID]Vote),
		openedAt:  openedAt,
	}
}

func (round Round) ID() RoundID {
	return round.id
}

func (round Round) SessionID() SessionID {
	return round.sessionID
}

func (round Round) Number() int {
	return round.number
}

func (round Round) State() RoundState {
	return round.state
}

func (round Round) VoteCount() int {
	return len(round.votes)
}

func (round Round) HasVoteFrom(discordUserID DiscordUserID) bool {
	_, ok := round.votes[discordUserID]

	return ok
}

func (round Round) VotedParticipantIDs() []DiscordUserID {
	discordUserIDs := make([]DiscordUserID, 0, len(round.votes))

	for discordUserID := range round.votes {
		discordUserIDs = append(discordUserIDs, discordUserID)
	}

	slices.SortFunc(discordUserIDs, func(left DiscordUserID, right DiscordUserID) int {
		return strings.Compare(string(left), string(right))
	})

	return discordUserIDs
}

func (round Round) NumericStatistics() (NumericStatistics, bool) {
	return round.numericStatistics, round.hasNumericStatistics
}

func (round Round) OpenedAt() time.Time {
	return round.openedAt
}

func (round Round) RevealedAt() time.Time {
	return round.revealedAt
}

func (round Round) ClosedAt() time.Time {
	return round.closedAt
}

func (round *Round) castVote(discordUserID DiscordUserID, estimate Estimate, castAt time.Time) {
	existing, exists := round.votes[discordUserID]

	if exists {
		existing.estimate = estimate
		existing.lastCastAt = castAt
		round.votes[discordUserID] = existing

		return
	}

	round.votes[discordUserID] = Vote{
		roundID:       round.id,
		discordUserID: discordUserID,
		estimate:      estimate,
		firstCastAt:   castAt,
		lastCastAt:    castAt,
	}
}

func (round *Round) reveal(statistics NumericStatistics, revealedAt time.Time) {
	round.state = RoundStateRevealed
	round.numericStatistics = statistics
	round.hasNumericStatistics = true
	round.revealedAt = revealedAt
}

func (round *Round) close(closedAt time.Time) {
	round.state = RoundStateClosed
	round.closedAt = closedAt
}
