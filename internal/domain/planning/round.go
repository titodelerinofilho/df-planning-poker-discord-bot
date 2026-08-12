package planning

import "time"

type RoundID string

type RoundState string

const (
	RoundStateOpen     RoundState = "OPEN"
	RoundStateReady    RoundState = "READY"
	RoundStateRevealed RoundState = "REVEALED"
	RoundStateClosed   RoundState = "CLOSED"
)

type Round struct {
	id        RoundID
	sessionID SessionID
	number    int
	state     RoundState
	openedAt  time.Time
}

func newFirstRound(id RoundID, sessionID SessionID, openedAt time.Time) Round {
	return Round{
		id:        id,
		sessionID: sessionID,
		number:    1,
		state:     RoundStateOpen,
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

func (round Round) OpenedAt() time.Time {
	return round.openedAt
}
