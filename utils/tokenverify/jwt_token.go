package tokenverify

import (
	"errors"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

var (
	ErrTokenMalformed = apperr.New("token_malformed_error",
		apperr.WithTextTranslate(translator.Translate{
			translator.RU: "Неверный токен",
			translator.EN: "Token malformed",
		}),
		apperr.WithCode(code.Unauthenticated),
	)
	ErrTokenNotValidYet = apperr.New("token_not_valid_yet",
		apperr.WithTextTranslate(translator.Translate{
			translator.RU: "Недействительный токен",
			translator.EN: "Token not valid yet",
		}),
		apperr.WithCode(code.Unauthenticated),
	)
	ErrTokenUnknown = apperr.New("token_unknown",
		apperr.WithTextTranslate(translator.Translate{
			translator.RU: "Неизвестный токен",
			translator.EN: "Token unknown",
		}),
		apperr.WithCode(code.Unauthenticated),
	)
	ErrTokenExpired = apperr.New("token_expired",
		apperr.WithTextTranslate(translator.Translate{
			translator.RU: "Срок действия токена истек",
			translator.EN: "Token expired",
		}),
		apperr.WithCode(code.Unauthenticated),
	)
	ErrTokenNotExist = apperr.New("token_not_exist",
		apperr.WithTextTranslate(translator.Translate{
			translator.RU: "Токен не найден",
			translator.EN: "Token not exist",
		}),
		apperr.WithCode(code.Unauthenticated),
	)
	ErrTokenKicked = apperr.New("token_kicked",
		apperr.WithTextTranslate(translator.Translate{
			translator.RU: "Токен удален",
			translator.EN: "Token kicked",
		}),
		apperr.WithCode(code.Unauthenticated),
	)
)

const minutesBefore = 5

type Claims struct {
	UserID   int
	DeviceID int // login Device
	jwt.RegisteredClaims
}

func BuildClaims(userID int, DeviceID int, ttl time.Duration) Claims {
	now := time.Now()
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
	token, err := jwt.ParseWithClaims(tokensString, &Claims{}, secretFunc)
	if err == nil {
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
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
