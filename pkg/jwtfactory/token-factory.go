package jwtfactory

import (
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"time"
)

type TokenFactory struct {
	tokenAuth           *jwtauth.JWTAuth
	tokenExpirationTime time.Duration
}

func New(tokenAuth *jwtauth.JWTAuth, tokenExpirationTime time.Duration) *TokenFactory {
	return &TokenFactory{
		tokenAuth:           tokenAuth,
		tokenExpirationTime: tokenExpirationTime,
	}
}

func (tf *TokenFactory) Generate(extraClaims map[string]string) (string, error) {
	timeNow := time.Now()
	claims := map[string]any{
		"exp": timeNow.Add(tf.tokenExpirationTime).Unix(),
		"iat": timeNow.Unix(),
	}
	for k, v := range extraClaims {
		claims[k] = v
	}
	_, tokenString, err := tf.tokenAuth.Encode(claims)
	return tokenString, fmt.Errorf("failed to generate token: %w", err)
}
