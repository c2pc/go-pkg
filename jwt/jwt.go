package jwt

import (
	"github.com/c2pc/go-pkg/apperr"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"strings"
	"time"
)

const AuthUserKey = "authUser"
const authorizationHeader = "Authorization"

var (
	ErrEmptyAuthHeader      = apperr.New("empty_auth_header")
	ErrInvalidAuthHeader    = apperr.New("invalid_auth_header")
	ErrEmptyToken           = apperr.New("empty_token")
	ErrInvalidToken         = apperr.New("invalid_token")
	ErrTokenParseError      = apperr.New("token_parse_error")
	ErrErrorToSigningString = apperr.New("error_to_signing_string")
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

type TokenClaims struct {
	Id           int    `json:"id"`
	Role         string `json:"role"`
	DepartmentId int    `json:"department_id"`
	jwt.RegisteredClaims
}

type User struct {
	Id           int    `json:"id"`
	Role         string `json:"role"`
	DepartmentId int    `json:"department_id"`
}

func ParseAuthHeader(c *gin.Context) (string, error) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return "", ErrEmptyAuthHeader
	}
	headerParts := strings.Split(header, " ")

	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", ErrInvalidAuthHeader
	}
	if len(headerParts[1]) == 0 {
		return "", ErrEmptyToken
	}

	return headerParts[1], nil
}

func (j *JWT) ParseToken(token string) (*User, error) {
	bearerToken, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if jwt.GetSigningMethod(j.Algo) != token.Method {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return j.Key, nil
	})
	if err != nil {
		return nil, ErrTokenParseError.WithError(err)
	}

	if claims, ok := bearerToken.Claims.(*TokenClaims); ok && bearerToken.Valid {
		return &User{
			Id:           claims.Id,
			Role:         claims.Role,
			DepartmentId: claims.DepartmentId,
		}, nil
	} else {
		return nil, ErrInvalidToken
	}
}

func (j *JWT) GenerateToken(u User) (string, float64, error) {
	claims := TokenClaims{
		u.Id,
		u.Role,
		u.DepartmentId,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.Duration)),
			ID:        strconv.FormatInt(time.Now().Unix(), 10),
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod(j.Algo), claims)
	tokenString, err := token.SignedString(j.Key)
	if err != nil {
		return "", 0, ErrErrorToSigningString.WithError(err)
	}

	return tokenString, j.Duration.Seconds(), nil
}
