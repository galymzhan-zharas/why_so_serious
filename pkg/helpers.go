package pkg

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
)

func ParseJWT(jwtStr string, secret string) (int64, string, error) {
	token, err := jwt.Parse(jwtStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Invalid sign method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, "", err
	}

	if !token.Valid {
		return 0, "", jwt.ErrTokenExpired
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", jwt.ErrTokenInvalidClaims
	}

	userID := int64(claims["userId"].(float64))
	userAgent := claims["userAgent"].(string)

	return userID, userAgent, nil
}
