package token

import "time"

//Maker is an interface for managing tokens
type Maker interface {
	//CreateToken create a new token for specific national id and duration
	CreateToken(nationalID string, duration time.Duration) (string, *Payload, error)

	//VerfifyToken checks if the token is valid or not
	VerifyToken(token string) (*Payload, error)
}
