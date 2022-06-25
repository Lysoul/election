package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"election/util"

	"github.com/stretchr/testify/require"
)

func TestCreateCandidate(t *testing.T) {
	CreateCandidate(t)
}

func TestGetCandidate(t *testing.T) {
	candidate1 := CreateCandidate(t)

	candidate2, err := testQueries.GetCandidate(context.Background(), candidate1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, candidate2)

	require.Equal(t, candidate1.ID, candidate2.ID)
	require.Equal(t, candidate1.Name, candidate2.Name)
	require.Equal(t, candidate1.Dob, candidate2.Dob)
	require.Equal(t, candidate1.BioLink, candidate2.BioLink)
	require.Equal(t, candidate1.ImageUrl, candidate2.ImageUrl)
	require.Equal(t, candidate1.Policy, candidate2.Policy)
	require.Equal(t, candidate1.VoteCount, candidate2.VoteCount)
	require.WithinDuration(t, candidate1.CreateAt, candidate2.CreateAt, time.Second)
}

func TestUpdateCandidate(t *testing.T) {
	candidate := CreateCandidate(t)
	arg := UpdateCandidateParams{
		ID:       candidate.ID,
		Name:     util.RandomName(),
		Dob:      util.RandomString(10),
		BioLink:  util.RandomBioLink(),
		ImageUrl: util.RandomImageLink(),
		Policy:   util.RandomString(15),
	}
	updatedCandidate, err := testQueries.UpdateCandidate(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedCandidate)

	require.Equal(t, arg.ID, updatedCandidate.ID)
	require.Equal(t, arg.Name, updatedCandidate.Name)
	require.Equal(t, arg.Dob, updatedCandidate.Dob)
	require.Equal(t, arg.BioLink, updatedCandidate.BioLink)
	require.Equal(t, arg.ImageUrl, updatedCandidate.ImageUrl)
	require.Equal(t, arg.Policy, updatedCandidate.Policy)
}

func TestDeleteCandidate(t *testing.T) {
	candidate1 := CreateCandidate(t)
	err := testQueries.DeleteCandidate(context.Background(), candidate1.ID)
	require.NoError(t, err)

	candidate2, err := testQueries.GetCandidate(context.Background(), candidate1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, candidate2)
}

func TestListCandidates(t *testing.T) {
	for i := 0; i < 10; i++ {
		CreateCandidate(t)
	}

	arg := ListCandidatesParams{
		Limit:  5,
		Offset: 5,
	}

	candidates, err := testQueries.ListCandidates(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, candidates, 5)

	for _, candidate := range candidates {
		require.NotEmpty(t, candidate)
	}
}

func TestListCandidatesResult(t *testing.T) {
	for i := 0; i < 3; i++ {
		CreateCandidate(t)
	}

	candidates, err := testQueries.ListCandidatesResult(context.Background())
	require.NoError(t, err)

	for _, candidate := range candidates {
		require.NotEmpty(t, candidate)
	}
}

func CreateCandidate(t *testing.T) Candidate {
	arg := CreateCandidateParams{
		Name:      util.RandomName(),
		Dob:       util.RandomDob(),
		BioLink:   util.RandomBioLink(),
		ImageUrl:  util.RandomImageLink(),
		Policy:    util.RandomString(15),
		VoteCount: 0,
	}

	candidate, err := testQueries.CreateCandidate(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, candidate)

	require.Equal(t, arg.Name, candidate.Name)
	require.Equal(t, arg.Dob, candidate.Dob)
	require.Equal(t, arg.BioLink, candidate.BioLink)
	require.Equal(t, arg.ImageUrl, candidate.ImageUrl)
	require.Equal(t, arg.Policy, candidate.Policy)
	require.Equal(t, arg.VoteCount, candidate.VoteCount)
	require.NotZero(t, candidate.ID)
	require.NotZero(t, candidate.CreateAt)
	return candidate
}
