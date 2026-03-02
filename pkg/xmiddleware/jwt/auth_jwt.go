package jwt

import (
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// CustomerInfo the content of the data returned to the frontend
type CustomerInfo struct {
	UserId   string `json:"user_id"`  // ID
	Username string `json:"username"` // username
}

// CustomClaims jwt claims definition
type CustomClaims struct {
	*jwt.StandardClaims
	TokenType string
	CustomerInfo
}

// Token the content of the token returned to the client
type Token struct {
	AccessToken string `json:"access_token"` // token
	ExpiresIn   int64  `json:"expires_in"`   // expiration time
	CustomerInfo
}

// CreateToken get a jwt token
func CreateToken(info CustomerInfo, jwtSecret []byte) (*Token, error) {
	expiresAt := time.Now().Add(time.Minute * 60).Unix()

	claims := &CustomClaims{
		&jwt.StandardClaims{

			ExpiresAt: expiresAt,
			Issuer:    "GasPc",
		},
		"level1",
		info,
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtTokenStr, err := tokenClaims.SignedString(jwtSecret)

	var token Token
	if err != nil {
		return &token, err
	}

	token.AccessToken = jwtTokenStr
	token.ExpiresIn = expiresAt
	token.Username = info.Username
	token.UserId = info.UserId

	return &token, nil
}

// ParseToken token verify
func ParseToken(tokenString string, jwtSecret []byte) (*CustomClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*CustomClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}

func GinParse(c *gin.Context, jwtSecret []byte) (cla CustomClaims, err error) {
	auth := c.GetHeader("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")

	if token == "" {
		err = errors.New("parsing Bearer Token failed")
		return
	}

	tokenClaims, err := jwt.ParseWithClaims(token, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*CustomClaims); ok && tokenClaims.Valid {
			return *claims, nil
		}
	}

	return
}
