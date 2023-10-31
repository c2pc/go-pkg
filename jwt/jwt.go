package jwt

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/c2pc/go-pkg/apperr"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

const AuthUserKey = "authUser"
const authorizationHeader = "Authorization"

var ErrUnauthorized = apperr.NewMethod("", "auth").
	WithTitleTranslate(apperr.Translate{"ru": "Попытка авторизации"})

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
	Token string
}

func (j *JWT) Authenticate(c *gin.Context) {
	token, err := ParseAuthHeader(c)
	if err != nil {
		apperr.HTTPResponse(c, ErrUnauthorized.Combine(apperr.ErrUnauthorized.WithError(err)))
		return
	}

	tokenClaims, err := j.ParseToken(token)
	if err != nil {
		apperr.HTTPResponse(c, ErrUnauthorized.Combine(apperr.ErrUnauthorized.WithError(err)))
		return
	}

	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), AuthUserKey, tokenClaims))

	c.Next()
}

func ParseAuthHeader(c *gin.Context) (string, error) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return "", errors.New("empty auth header")
	}
	headerParts := strings.Split(header, " ")

	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", errors.New("invalid auth header")
	}
	if len(headerParts[1]) == 0 {
		return "", errors.New("token is empty")
	}

	return headerParts[1], nil
}

func (j *JWT) ParseToken(token string) (*Claims, error) {
	bearerToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(j.Algo) != token.Method {
			return nil, apperr.ErrInternal.WithError(errors.New("invalid signing method"))
		}
		return j.Key, nil
	})
	if err != nil {
		return nil, ErrUnauthorized.Combine(apperr.ErrUnauthorized.WithError(err))
	}

	if !bearerToken.Valid {
		return nil, ErrUnauthorized.Combine(apperr.ErrUnauthorized.WithError(errors.New("invalid token")))
	}

	mapClaims := bearerToken.Claims.(jwt.MapClaims)

	return &Claims{
		Id:    int(mapClaims["id"].(float64)),
		Login: mapClaims["l"].(string),
		Role:  mapClaims["n"].(string),
		Token: token,
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
