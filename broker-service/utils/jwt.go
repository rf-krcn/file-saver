package utils

import (
	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte("your-secret-key")

// ValidateToken validates a JWT token.
func ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}
