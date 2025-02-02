package jwtfactory

import (
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

func (tf *TokenFactory) Generate(login string) (string, error) {
	timeNow := time.Now()
	claims := map[string]any{
		"login": login,
		"exp":   timeNow.Add(tf.tokenExpirationTime).Unix(),
		"iat":   timeNow.Unix(),
	}
	_, tokenString, err := tf.tokenAuth.Encode(claims)
	return tokenString, err
}
