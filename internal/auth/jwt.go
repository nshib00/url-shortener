package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID int, secretKey string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": userID,
			"exp": time.Now().Add(24 * time.Hour).Unix(),
			"iat": time.Now().Unix(),
		},
	)
	tokenStr, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func ValidateToken(tokenStr string, secretKey string) (int, error) {
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	}
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwt.MapClaims{},
		keyFunc,
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid claims")
	}
	userID := int((*claims)["sub"].(float64))
	return userID, nil
}
