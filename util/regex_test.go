package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDateOfBirthRegex(t *testing.T) {
	s := "August 8, 2011"
	isDobRegex := IsDateOfBirth(s)
	require.True(t, isDobRegex)
}

func TestInvalidDaysDobRegex(t *testing.T) {
	s := "August invalid, 2011"
	isDobRegex := IsDateOfBirth(s)
	require.False(t, isDobRegex)
}

func TestInvalidDobRegex(t *testing.T) {
	s := "Invalid"
	isDobRegex := IsDateOfBirth(s)
	require.False(t, isDobRegex)
}
