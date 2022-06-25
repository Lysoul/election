package db

import (
	"context"
	"testing"
	"time"

	"election/util"

	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	CreateUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := CreateUser(t)

	user2, err := testQueries.GetUser(context.Background(), user1.NationalID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.NationalID, user2.NationalID)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Permission, user2.Permission)
	require.Equal(t, user1.HasVoted, user2.HasVoted)

	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
	require.WithinDuration(t, user1.CreateAt, user2.CreateAt, time.Second)
}

func CreateUser(t *testing.T) User {
	hasedPassword := "secret"

	permission := make([]string, 1)
	permission[0] = util.Vote

	arg := CreateUserParams{
		NationalID:     util.RandomString(13),
		HashedPassword: hasedPassword,
		FullName:       util.RandomName(),
		Email:          util.RandomEmail(),
		Permission:     permission,
		HasVoted:       false,
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.NationalID, user.NationalID)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Permission, user.Permission)
	require.Equal(t, arg.HasVoted, user.HasVoted)
	require.NotZero(t, user.CreateAt)
	return user
}
