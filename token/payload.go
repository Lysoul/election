package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

//Difference type of errors return by the VerifyToken function
var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is invalid")
)

//Payload contains the payload data of the token
type Payload struct {
	ID          uuid.UUID `json:"id"`
	NationalID  string    `json:"national_id"`
	Permissions []string  `json:"permissions"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiredAt   time.Time `json:"expired_at"`
}

//NewPayload creates a new token with specific national id and duration
func NewPayload(nationalID string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:         tokenID,
		NationalID: nationalID,
		IssuedAt:   time.Now(),
		ExpiredAt:  time.Now().Add(duration),
	}

	return payload, nil
}

//Valid checks if the token payload is valid or not
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}

	return nil
}
