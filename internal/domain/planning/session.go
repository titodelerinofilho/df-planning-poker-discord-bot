package planning

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var ErrInvalidSession = errors.New("invalid planning session")

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
