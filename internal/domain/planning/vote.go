package planning

import "time"

type Vote struct {
	roundID       RoundID
	discordUserID DiscordUserID
	estimate      Estimate
	firstCastAt   time.Time
	lastCastAt    time.Time
}

func (vote Vote) RoundID() RoundID {
	return vote.roundID
}

func (vote Vote) DiscordUserID() DiscordUserID {
	return vote.discordUserID
}

func (vote Vote) Estimate() Estimate {
	return vote.estimate
}

func (vote Vote) FirstCastAt() time.Time {
	return vote.firstCastAt
}

func (vote Vote) LastCastAt() time.Time {
	return vote.lastCastAt
}
