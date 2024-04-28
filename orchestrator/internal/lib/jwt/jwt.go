package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrTokenExpired            = errors.New("token has expired")
)

func ValidateToken(tokenString, secret string) (int, error) {
	tokenFromString, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpectedSigningMethod
		}
		return []byte(secret), nil
	})
	if err != nil {
		return -1, err
	}

	var (
		claims jwt.MapClaims
		ok     bool
	)
	if claims, ok = tokenFromString.Claims.(jwt.MapClaims); !ok {
		return -1, err
	}

	if int64(claims["exp"].(float64))-time.Now().Unix() <= 0 {
		return -1, ErrTokenExpired
	}
	return int(claims["uid"].(float64)), nil
}
