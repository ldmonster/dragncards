package auth

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret     = []byte("dev-secret-change-me")
	revokedTokens = struct {
		m map[string]struct{}
		sync.RWMutex
	}{m: map[string]struct{}{}}
)

func init() {
	if s := os.Getenv("AUTH_TOKEN_SECRET"); s != "" {
		jwtSecret = []byte(s)
	}
}

type jwtClaims struct {
	jwt.RegisteredClaims
}

func SetSecret(secret string) {
	jwtSecret = []byte(secret)
}

func GenerateToken(subject string, ttl time.Duration) (string, error) {
	if subject == "" || ttl <= 0 {
		return "", errors.New("invalid token generation parameters")
	}

	claims := jwtClaims{RegisteredClaims: jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return signed, nil
}

func ParseToken(tokenString string) (string, error) {
	if tokenString == "" {
		return "", errors.New("token is empty")
	}

	if IsTokenRevoked(tokenString) {
		return "", errors.New("token revoked")
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}

	if claims.Subject == "" {
		return "", errors.New("missing subject")
	}

	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return "", errors.New("token expired")
	}

	return claims.Subject, nil
}

func RevokeToken(tokenString string) {
	revokedTokens.Lock()
	defer revokedTokens.Unlock()
	revokedTokens.m[tokenString] = struct{}{}
}

func IsTokenRevoked(tokenString string) bool {
	revokedTokens.RLock()
	defer revokedTokens.RUnlock()
	_, ok := revokedTokens.m[tokenString]
	return ok
}
