package token

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Token struct
type Token struct {
	secret string
}

var (
	// We are going to expire all tokens at zero hour (Brasilia Standard Time) of the day of the election.
	expirationDate = time.Date(2020, 11, 15, 0, 0, 0, 0, time.UTC).Add(-3 * time.Hour).Unix() // 3 is the difference between UTC and BST.
)

// New returns a new token service
func New(secret string) *Token {
	return &Token{
		secret: secret,
	}
}

// GetToken returns a new token
func (t *Token) GetToken(email string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   expirationDate,
	})
	return token.SignedString([]byte(t.secret))
}

// IsValid checks if token is valid
func (t *Token) IsValid(auhtorization string) bool {
	token, err := jwt.Parse(auhtorization, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("error on validating auth token")
		}
		return []byte(t.secret), nil
	})
	if err != nil {
		return false
	}
	return token.Valid
}

// GetClaims transforms the token string into a map with its claims
func GetClaims(auhtorization string) (map[string]string, error) {
	token, _ := jwt.Parse(auhtorization, func(token *jwt.Token) (interface{}, error) {
		return []byte(""), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		claimsMap := make(map[string]string)
		for key, value := range claims {
			if key != "exp" {
				claimsMap[key] = value.(string)
			}
		}
		return claimsMap, nil
	}
	return nil, fmt.Errorf("could not get claims")

}
