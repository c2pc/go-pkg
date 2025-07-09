package tokenverify

import (
	"errors"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/i18n"
	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrTokenMalformed   = apperr.New("token_malformed_error", apperr.WithTextTranslate(i18n.ErrTokenMalformed), apperr.WithCode(code.Unauthenticated))
	ErrTokenNotValidYet = apperr.New("token_not_valid_yet", apperr.WithTextTranslate(i18n.ErrTokenNotValidYet), apperr.WithCode(code.Unauthenticated))
	ErrTokenUnknown     = apperr.New("token_unknown", apperr.WithTextTranslate(i18n.ErrTokenUnknown), apperr.WithCode(code.Unauthenticated))
	ErrTokenExpired     = apperr.New("token_expired", apperr.WithTextTranslate(i18n.ErrTokenExpired), apperr.WithCode(code.Unauthenticated))
	ErrTokenNotExist    = apperr.New("token_not_exist", apperr.WithTextTranslate(i18n.ErrTokenNotExist), apperr.WithCode(code.Unauthenticated))
	ErrTokenKicked      = apperr.New("token_kicked", apperr.WithTextTranslate(i18n.ErrTokenKicked), apperr.WithCode(code.Unauthenticated))
)

const minutesBefore = 5

type Claims struct {
	UserID   int
	DeviceID int // login Device
	jwt.RegisteredClaims
}

func BuildClaims(userID int, DeviceID int, ttl time.Duration) Claims {
	now := time.Now().UTC()
	before := now.Add(-time.Minute * time.Duration(minutesBefore))
	return Claims{
		UserID:   userID,
		DeviceID: DeviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)), // Expiration time
			IssuedAt:  jwt.NewNumericDate(now),          // Issuing time
			NotBefore: jwt.NewNumericDate(before),       // Begin Effective time
		},
	}
}

func GetClaimFromToken(tokensString string, secretFunc jwt.Keyfunc) (*Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokensString, &claims, secretFunc)
	if err == nil {
		if claims, ok := token.Claims.(*Claims); ok {
			if token.Valid {
				return claims, nil
			}
			return claims, ErrTokenUnknown
		}
		return &claims, ErrTokenUnknown
	}

	var ve *jwt.ValidationError
	if errors.As(err, &ve) {
		return &claims, mapValidationError(ve)
	}

	return &claims, ErrTokenUnknown
}

func mapValidationError(ve *jwt.ValidationError) error {
	if ve.Errors&jwt.ValidationErrorMalformed != 0 {
		return ErrTokenMalformed
	} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
		return ErrTokenExpired
	} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
		return ErrTokenNotValidYet
	}
	return ErrTokenUnknown
}

func Secret(secret string) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}
}
