package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"trenerbot/internal/domain"
)

// TokenService issues short-lived access JWTs. Refresh tokens are opaque random
// strings persisted (hashed) in the store, handled by the service layer.
type TokenService struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenService(secret string, accessTTL time.Duration) *TokenService {
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}
	return &TokenService{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: 30 * 24 * time.Hour,
	}
}

// RefreshTTL is the lifetime used when persisting refresh tokens.
func (t *TokenService) RefreshTTL() time.Duration { return t.refreshTTL }

type Claims struct {
	UserID int64       `json:"uid"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// Generate issues an access token. Kept for backward compatibility (bot flow).
func (t *TokenService) Generate(userID int64, role domain.Role) (string, error) {
	return t.GenerateAccess(userID, role)
}

func (t *TokenService) GenerateAccess(userID int64, role domain.Role) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.accessTTL)),
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
