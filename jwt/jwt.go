package jwt

import (
	"crypto/rand"
	"errors"
	"github.com/c2pc/go-pkg/apperr"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type JWT struct {
	Key      []byte
	Duration time.Duration
	Algo     string
}

func NewJWT(signingKey string, accessTokenTTL time.Duration, signingAlgorithm string) *JWT {
	return &JWT{
		Key:      []byte(signingKey),
		Duration: accessTokenTTL,
		Algo:     signingAlgorithm,
	}
}

type Claims struct {
	Id    int
	Login string
	Role  string
}

func (j *JWT) ParseToken(token string) (*Claims, error) {
	bearerToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(j.Algo) != token.Method {
			return nil, apperr.ErrInternal.WithError(errors.New("invalid signing method"))
		}
		return j.Key, nil
	})
	if err != nil {
		return nil, apperr.ErrUnauthenticated.WithError(err)
	}

	if !bearerToken.Valid {
		return nil, apperr.ErrUnauthenticated.WithError(errors.New("invalid token"))
	}

	mapClaims := bearerToken.Claims.(jwt.MapClaims)

	return &Claims{
		Id:    int(mapClaims["id"].(float64)),
		Login: mapClaims["l"].(string),
		Role:  mapClaims["n"].(string),
	}, nil
}

func (j *JWT) GenerateToken(u Claims) (string, float64, error) {
	token := jwt.New(jwt.GetSigningMethod(j.Algo))
	claims := token.Claims.(jwt.MapClaims)

	bits := make([]byte, 12)
	_, err := rand.Read(bits)
	if err != nil {
		return "", 0, err
	}

	claims["id"] = u.Id
	claims["l"] = u.Login
	claims["n"] = u.Role
	claims["exp"] = time.Now().Add(j.Duration).Unix()
	claims["tid"] = bits

	tokenString, err := token.SignedString(j.Key)
	return tokenString, j.Duration.Seconds(), err
}
