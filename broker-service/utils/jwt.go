package utils

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte("your-secret-key")

func ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
}

func DecodeJWT(tokenString string) (map[string]interface{}, error) {

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, jwt.ErrInvalidKey
	}

	decodedPayload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var payloadMap map[string]interface{}
	err = json.Unmarshal(decodedPayload, &payloadMap)
	if err != nil {
		return nil, err
	}

	return payloadMap, nil
}
