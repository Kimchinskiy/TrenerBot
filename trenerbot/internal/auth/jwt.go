package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"trenerbot/internal/domain"
)

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string, ttl time.Duration) *TokenService {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &TokenService{secret: []byte(secret), ttl: ttl}
}

type Claims struct {
	UserID int64      `json:"uid"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

func (t *TokenService) Generate(userID int64, role domain.Role) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
			Subject:   string(role),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(t.secret)
}

func (t *TokenService) Parse(token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(c *jwt.Token) (interface{}, error) {
		if _, ok := c.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return t.secret, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}
