package planning

import (
	"time"
)

type RevealedVote struct {
	discordUserID DiscordUserID
	displayName   string
	estimate      Estimate
}

type RoundRevealResult struct {
	roundID           RoundID
	roundNumber       int
	votes             []RevealedVote
	numericStatistics NumericStatistics
	revealedAt        time.Time
}

func newRoundRevealResult(
	roundID RoundID,
	roundNumber int,
	votes []RevealedVote,
	numericStatistics NumericStatistics,
	revealedAt time.Time,
) RoundRevealResult {
	votesCopy := make([]RevealedVote, len(votes))
	copy(votesCopy, votes)

	return RoundRevealResult{
		roundID:           roundID,
		roundNumber:       roundNumber,
		votes:             votesCopy,
		numericStatistics: numericStatistics,
		revealedAt:        revealedAt,
	}
}

func (result RoundRevealResult) RoundID() RoundID {
	return result.roundID
}

func (result RoundRevealResult) RoundNumber() int {
	return result.roundNumber
}

func (result RoundRevealResult) Votes() []RevealedVote {
	votes := make([]RevealedVote, len(result.votes))
	copy(votes, result.votes)

	return votes
}

func (result RoundRevealResult) NumericStatistics() NumericStatistics {
	return result.numericStatistics
}

func (result RoundRevealResult) RevealedAt() time.Time {
	return result.revealedAt
}

func (vote RevealedVote) DiscordUserID() DiscordUserID {
	return vote.discordUserID
}

func (vote RevealedVote) DisplayName() string {
	return vote.displayName
}

func (vote RevealedVote) Estimate() Estimate {
	return vote.estimate
}

func revealedVotesByParticipantOrder(participants []Participant, votes map[DiscordUserID]Vote) []RevealedVote {
	revealedVotes := make([]RevealedVote, 0, len(votes))

	for _, participant := range participants {
		vote, ok := votes[participant.discordUserID]

		if !ok {
			continue
		}

		revealedVotes = append(revealedVotes, RevealedVote{
			discordUserID: participant.discordUserID,
			displayName:   participant.displayName,
			estimate:      vote.estimate,
		})
	}

	return revealedVotes
}
