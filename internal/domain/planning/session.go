package planning

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
)

var ErrInvalidSession = errors.New("invalid planning session")
var ErrSessionNotOpen = errors.New("planning session is not open")
var ErrParticipantNotFound = errors.New("participant not found")
var ErrAlreadyParticipant = errors.New("user already participates")
var ErrInvalidParticipant = errors.New("invalid participant")
var ErrNotEnoughParticipants = errors.New("not enough participants")
var ErrInvalidRound = errors.New("invalid round")
var ErrVoteNotAllowed = errors.New("vote is not allowed")

type SessionState string

const (
	SessionStateJoining       SessionState = "JOINING"
	SessionStateVoting        SessionState = "VOTING"
	SessionStateReadyToReveal SessionState = "READY_TO_REVEAL"
	SessionStateRevealed      SessionState = "REVEALED"
	SessionStateDiscussing    SessionState = "DISCUSSING"
	SessionStateCompleted     SessionState = "COMPLETED"
	SessionStateCancelled     SessionState = "CANCELLED"
	SessionStateExpired       SessionState = "EXPIRED"
)

type SessionID string
type GuildID string
type ChannelID string
type ThreadID string
type MessageID string
type DiscordUserID string

type Task struct {
	URL   string
	Title string
}

type Participant struct {
	sessionID     SessionID
	discordUserID DiscordUserID
	displayName   string
	active        bool
	joinedAt      time.Time
	leftAt        time.Time
}

type NewSessionInput struct {
	ID            SessionID
	GuildID       GuildID
	ChannelID     ChannelID
	ThreadID      ThreadID
	MessageID     MessageID
	Task          Task
	CreatorID     DiscordUserID
	FacilitatorID DiscordUserID
	Scale         Scale
	CreatedAt     time.Time
	ExpiresAt     time.Time
}

type Session struct {
	id                 SessionID
	guildID            GuildID
	channelID          ChannelID
	threadID           ThreadID
	messageID          MessageID
	task               Task
	creatorID          DiscordUserID
	facilitatorID      DiscordUserID
	scale              Scale
	participants       map[DiscordUserID]Participant
	currentRound       Round
	hasCurrentRound    bool
	state              SessionState
	currentRoundNumber int
	createdAt          time.Time
	expiresAt          time.Time
}

func NewSession(input NewSessionInput) (Session, error) {
	input.Task.URL = strings.TrimSpace(input.Task.URL)
	input.Task.Title = strings.TrimSpace(input.Task.Title)

	err := validateNewSessionInput(input)

	if err != nil {
		return Session{}, err
	}

	return Session{
		id:                 input.ID,
		guildID:            input.GuildID,
		channelID:          input.ChannelID,
		threadID:           input.ThreadID,
		messageID:          input.MessageID,
		task:               input.Task,
		creatorID:          input.CreatorID,
		facilitatorID:      input.FacilitatorID,
		scale:              input.Scale,
		participants:       make(map[DiscordUserID]Participant),
		state:              SessionStateJoining,
		currentRoundNumber: 0,
		createdAt:          input.CreatedAt,
		expiresAt:          input.ExpiresAt,
	}, nil
}

func (session Session) ID() SessionID {
	return session.id
}

func (session Session) GuildID() GuildID {
	return session.guildID
}

func (session Session) ChannelID() ChannelID {
	return session.channelID
}

func (session Session) ThreadID() ThreadID {
	return session.threadID
}

func (session Session) MessageID() MessageID {
	return session.messageID
}

func (session Session) Task() Task {
	return session.task
}

func (session Session) CreatorID() DiscordUserID {
	return session.creatorID
}

func (session Session) FacilitatorID() DiscordUserID {
	return session.facilitatorID
}

func (session Session) Scale() Scale {
	return session.scale
}

func (session Session) Participants() []Participant {
	participants := make([]Participant, 0, len(session.participants))

	for _, participant := range session.participants {
		participants = append(participants, participant)
	}

	slices.SortFunc(participants, compareParticipantsByJoinTime)

	return participants
}

func (session Session) ActiveParticipants() []Participant {
	participants := make([]Participant, 0, len(session.participants))

	for _, participant := range session.participants {
		if participant.active {
			participants = append(participants, participant)
		}
	}

	slices.SortFunc(participants, compareParticipantsByJoinTime)

	return participants
}

func (session Session) CurrentRound() (Round, bool) {
	return session.currentRound, session.hasCurrentRound
}

func (session Session) State() SessionState {
	return session.state
}

func (session Session) CurrentRoundNumber() int {
	return session.currentRoundNumber
}

func (session Session) CreatedAt() time.Time {
	return session.createdAt
}

func (session Session) ExpiresAt() time.Time {
	return session.expiresAt
}

func (session *Session) JoinParticipant(discordUserID DiscordUserID, displayName string, joinedAt time.Time) error {
	if session.state != SessionStateJoining {
		return fmt.Errorf("%w: participants can only join while session is joining", ErrSessionNotOpen)
	}

	participant, err := newParticipant(session.id, discordUserID, displayName, joinedAt)

	if err != nil {
		return err
	}

	existing, exists := session.participants[participant.discordUserID]

	if exists && existing.active {
		return fmt.Errorf("%w: %s", ErrAlreadyParticipant, participant.discordUserID)
	}

	session.participants[participant.discordUserID] = participant

	return nil
}

func (session *Session) LeaveParticipant(discordUserID DiscordUserID, leftAt time.Time) error {
	if session.state != SessionStateJoining {
		return fmt.Errorf("%w: participants can only leave while session is joining", ErrSessionNotOpen)
	}

	if discordUserID == "" {
		return fmt.Errorf("%w: discord user id is required", ErrInvalidParticipant)
	}

	if leftAt.IsZero() {
		return fmt.Errorf("%w: left at is required", ErrInvalidParticipant)
	}

	participant, exists := session.participants[discordUserID]

	if !exists || !participant.active {
		return fmt.Errorf("%w: %s", ErrParticipantNotFound, discordUserID)
	}

	participant.active = false
	participant.leftAt = leftAt
	session.participants[discordUserID] = participant

	return nil
}

func (session *Session) CloseParticipants(minimumParticipants int, roundID RoundID, openedAt time.Time) error {
	if session.state != SessionStateJoining {
		return fmt.Errorf("%w: participants can only close while session is joining", ErrSessionNotOpen)
	}

	if minimumParticipants < 1 {
		return fmt.Errorf("%w: minimum participants must be positive", ErrInvalidParticipant)
	}

	if roundID == "" {
		return fmt.Errorf("%w: round id is required", ErrInvalidRound)
	}

	if openedAt.IsZero() {
		return fmt.Errorf("%w: opened at is required", ErrInvalidRound)
	}

	activeParticipants := session.ActiveParticipants()

	if len(activeParticipants) < minimumParticipants {
		return fmt.Errorf(
			"%w: got %d active participants, want at least %d",
			ErrNotEnoughParticipants,
			len(activeParticipants),
			minimumParticipants,
		)
	}

	session.currentRound = newFirstRound(roundID, session.id, openedAt)
	session.hasCurrentRound = true
	session.currentRoundNumber = session.currentRound.number
	session.state = SessionStateVoting

	return nil
}

func (session *Session) CastVote(discordUserID DiscordUserID, estimate Estimate, castAt time.Time) error {
	if session.state != SessionStateVoting && session.state != SessionStateReadyToReveal {
		return fmt.Errorf("%w: session is not accepting votes", ErrVoteNotAllowed)
	}

	if !session.hasCurrentRound {
		return fmt.Errorf("%w: current round is required", ErrInvalidRound)
	}

	if session.currentRound.state != RoundStateOpen && session.currentRound.state != RoundStateReady {
		return fmt.Errorf("%w: round is not accepting votes", ErrVoteNotAllowed)
	}

	if castAt.IsZero() {
		return fmt.Errorf("%w: cast at is required", ErrVoteNotAllowed)
	}

	participant, exists := session.participants[discordUserID]

	if !exists || !participant.active {
		return fmt.Errorf("%w: %s", ErrParticipantNotFound, discordUserID)
	}

	err := session.scale.Validate(estimate)

	if err != nil {
		return err
	}

	session.currentRound.castVote(discordUserID, estimate, castAt)

	if session.currentRound.VoteCount() == len(session.ActiveParticipants()) {
		session.currentRound.state = RoundStateReady
		session.state = SessionStateReadyToReveal

		return nil
	}

	session.currentRound.state = RoundStateOpen
	session.state = SessionStateVoting

	return nil
}

func validateNewSessionInput(input NewSessionInput) error {
	if input.ID == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidSession)
	}

	if input.GuildID == "" {
		return fmt.Errorf("%w: guild id is required", ErrInvalidSession)
	}

	if input.ChannelID == "" {
		return fmt.Errorf("%w: channel id is required", ErrInvalidSession)
	}

	if input.ThreadID == "" {
		return fmt.Errorf("%w: thread id is required", ErrInvalidSession)
	}

	if input.MessageID == "" {
		return fmt.Errorf("%w: message id is required", ErrInvalidSession)
	}

	if input.CreatorID == "" {
		return fmt.Errorf("%w: creator id is required", ErrInvalidSession)
	}

	if input.FacilitatorID == "" {
		return fmt.Errorf("%w: facilitator id is required", ErrInvalidSession)
	}

	if input.CreatedAt.IsZero() {
		return fmt.Errorf("%w: created at is required", ErrInvalidSession)
	}

	if input.ExpiresAt.IsZero() {
		return fmt.Errorf("%w: expires at is required", ErrInvalidSession)
	}

	if !input.ExpiresAt.After(input.CreatedAt) {
		return fmt.Errorf("%w: expires at must be after created at", ErrInvalidSession)
	}

	if input.Task.URL == "" {
		return fmt.Errorf("%w: task url is required", ErrInvalidSession)
	}

	parsedURL, err := url.Parse(input.Task.URL)

	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("%w: task url is invalid", ErrInvalidSession)
	}

	if input.Task.Title == "" {
		return fmt.Errorf("%w: task title is required", ErrInvalidSession)
	}

	if len(input.Scale.Values()) == 0 {
		return fmt.Errorf("%w: scale is required", ErrInvalidSession)
	}

	return nil
}

func newParticipant(sessionID SessionID, discordUserID DiscordUserID, displayName string, joinedAt time.Time) (Participant, error) {
	displayName = strings.TrimSpace(displayName)

	if sessionID == "" {
		return Participant{}, fmt.Errorf("%w: session id is required", ErrInvalidParticipant)
	}

	if discordUserID == "" {
		return Participant{}, fmt.Errorf("%w: discord user id is required", ErrInvalidParticipant)
	}

	if displayName == "" {
		return Participant{}, fmt.Errorf("%w: display name is required", ErrInvalidParticipant)
	}

	if joinedAt.IsZero() {
		return Participant{}, fmt.Errorf("%w: joined at is required", ErrInvalidParticipant)
	}

	return Participant{
		sessionID:     sessionID,
		discordUserID: discordUserID,
		displayName:   displayName,
		active:        true,
		joinedAt:      joinedAt,
	}, nil
}

func (participant Participant) SessionID() SessionID {
	return participant.sessionID
}

func (participant Participant) DiscordUserID() DiscordUserID {
	return participant.discordUserID
}

func (participant Participant) DisplayName() string {
	return participant.displayName
}

func (participant Participant) Active() bool {
	return participant.active
}

func (participant Participant) JoinedAt() time.Time {
	return participant.joinedAt
}

func (participant Participant) LeftAt() time.Time {
	return participant.leftAt
}

func compareParticipantsByJoinTime(left Participant, right Participant) int {
	if left.joinedAt.Before(right.joinedAt) {
		return -1
	}

	if left.joinedAt.After(right.joinedAt) {
		return 1
	}

	return strings.Compare(string(left.discordUserID), string(right.discordUserID))
}
