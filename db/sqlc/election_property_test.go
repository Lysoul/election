package db

import (
	"context"
	"testing"

	"election/util"

	"github.com/stretchr/testify/require"
)

func TestGetHasClosedElection(t *testing.T) {
	hasClosedElection, err := testQueries.GetElectionProperty(context.Background(), util.ElectionClosed)
	require.NoError(t, err)
	require.NotEmpty(t, hasClosedElection)
}

func TestUpdateClosedElection(t *testing.T) {
	arg := UpdateElectionPropertyParams{
		Name:  util.ElectionClosed,
		Value: true,
	}

	hasClosedElection, err := testQueries.UpdateElectionProperty(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, hasClosedElection)
	require.Equal(t, arg.Name, hasClosedElection.Name)
	require.Equal(t, arg.Value, hasClosedElection.Value)

	arg.Value = false
	hasClosedElection, err = testQueries.UpdateElectionProperty(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, hasClosedElection)
}
