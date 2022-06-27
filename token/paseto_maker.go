package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

//PasetoMaker is a PASETO token maker
type PasetoMaker struct {
	paseto      *paseto.V2
	symmeticKey []byte
}

// CreateToken implements Maker
func (maker *PasetoMaker) CreateToken(nationalID string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(nationalID, duration)
	if err != nil {
		return "", payload, err
	}

	token, err := maker.paseto.Encrypt(maker.symmeticKey, payload, nil)
	return token, payload, err
}

// VerifyToken implements Maker
func (maker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}

	err := maker.paseto.Decrypt(token, maker.symmeticKey, payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}

//NewPasetoMaker crreate a new PasetoMaker
func NewPasetoMaker(symmeticKey string) (Maker, error) {
	if len(symmeticKey) < chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", chacha20poly1305.KeySize)
	}

	maker := &PasetoMaker{
		paseto:      paseto.NewV2(),
		symmeticKey: []byte(symmeticKey),
	}
	return maker, nil
}
