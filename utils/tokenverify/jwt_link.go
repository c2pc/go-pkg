package tokenverify

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const minutesLive = 15

type LinkClaims struct {
	Link string
	jwt.RegisteredClaims
}

func BuildLinkClaims(link string, ttl time.Duration) LinkClaims {
	now := time.Now().UTC()
	before := now.Add(-time.Minute * time.Duration(minutesBefore))
	return LinkClaims{
		Link: link,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)), // Expiration time
			IssuedAt:  jwt.NewNumericDate(now),          // Issuing time
			NotBefore: jwt.NewNumericDate(before),       // Begin Effective time
		},
	}
}

func GetLinkClaimFromToken(tokensString string, secretFunc jwt.Keyfunc) (*LinkClaims, error) {
	token, err := jwt.ParseWithClaims(tokensString, &LinkClaims{}, secretFunc)
	if err == nil {
		if claims, ok := token.Claims.(*LinkClaims); ok && token.Valid {
			return claims, nil
		}
		return nil, ErrTokenUnknown
	}

	var ve *jwt.ValidationError
	if errors.As(err, &ve) {
		return nil, mapValidationError(ve)
	}

	return nil, ErrTokenUnknown
}
