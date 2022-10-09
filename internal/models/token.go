package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/jwt"
)

type JWT struct {
	Auth     *jwtauth.JWTAuth
	ExpiryIn time.Duration
}

func NewJWT(alg string, secretKey string, expiryIn time.Duration) (*JWT, error) {
	fmt.Println("----------------------------------------------------------------")
	if len(secretKey) == 0 {
		return nil, errors.New("secret key empty")
	}

	if len(alg) == 0 {
		alg = "HS256"
	}

	jwt := &JWT{
		Auth:     jwtauth.New(alg, []byte(secretKey), nil),
		ExpiryIn: expiryIn,
	}

	return jwt, nil
}

func (jwt *JWT) Encode(claims map[string]interface{}) (string, error) {
	jwtauth.SetExpiryIn(claims, jwt.ExpiryIn)

	_, resToken, errToken := jwt.Auth.Encode(claims)
	if errToken != nil {
		return "", errors.New("Failed generation token: " + errToken.Error())
	}

	return resToken, nil
}

func (jwt *JWT) Parse(jwtToken string) (jwt.Token, error) {
	token, errDecode := jwt.Auth.Decode(jwtToken)
	if errDecode != nil {
		return nil, errors.New("Failed devode token from string: " + errDecode.Error())
	}

	return token, nil
}
