package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateVote(t *testing.T) {
	CreateVote(t)
}

func TestListVoteOrderByCandidate(t *testing.T) {
	for i := 0; i < 3; i++ {
		CreateVote(t)
	}

	listVotes, err := testQueries.ListVoteOrderByCandidate(context.Background())
	require.NoError(t, err)

	for _, listVote := range listVotes {
		require.NotEmpty(t, listVote)
	}
}

func CreateVote(t *testing.T) Vote {
	user := CreateUser(t)
	candidate := CreateCandidate(t)

	arg := CreateVoteParams{
		VoteNationalID: user.NationalID,
		CandidateID:    candidate.ID,
	}

	voted, err := testQueries.CreateVote(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, voted)

	require.Equal(t, arg.VoteNationalID, voted.VoteNationalID)
	require.Equal(t, arg.CandidateID, voted.CandidateID)

	candidate2, err := testQueries.GetCandidate(context.Background(), candidate.ID)
	require.NoError(t, err)
	require.Equal(t, candidate.VoteCount+1, candidate2.VoteCount)
	return voted
}
