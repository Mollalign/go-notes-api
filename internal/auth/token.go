package auth

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const userIDContextKey = "authenticated_user_id"

var ErrInvalidToken = errors.New("invalid access token")

type TokenManager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

func NewTokenManager(
	secret []byte,
	issuer string,
	audience string,
	ttl time.Duration,
) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must be at least 32 bytes")
	}

	if issuer == "" || audience == "" {
		return nil, errors.New("JWT issuer and audience are required")
	}

	if ttl <= 0 {
		return nil, errors.New("JWT expiration must be positive")
	}

	return &TokenManager{
		secret:   secret,
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
	}, nil
}

func (m *TokenManager) Issue(userID int64) (string, time.Time, error) {
	if userID <= 0 {
		return "", time.Time{}, errors.New("invalid user ID")
	}

	now := time.Now()
	expiresAt := now.Add(m.ttl)

	claims := jwt.RegisteredClaims{
		Issuer:   m.issuer,
		Subject:  strconv.FormatInt(userID, 10),
		Audience: jwt.ClaimStrings{m.audience},

		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	token.Header["typ"] = "JWT"

	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signedToken, expiresAt, nil
}

func (m *TokenManager) Verify(rawToken string) (int64, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}

			return m.secret, nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.audience),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)

	if err != nil || token == nil || !token.Valid {
		return 0, ErrInvalidToken
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID <= 0 {
		return 0, ErrInvalidToken
	}

	return userID, nil
}

func (m *TokenManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.Fields(header)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Authentication required",
			})
			return
		}

		userID, err := m.Verify(parts[1])
		if err != nil {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(401, gin.H{
				"error": "Invalid or expired access token",
			})
			return
		}

		c.Set(userIDContextKey, userID)

		c.Next()
	}
}

func UserIDFromContext(c *gin.Context) (int64, bool) {
	value, exists := c.Get(userIDContextKey)
	if !exists {
		return 0, false
	}

	userID, ok := value.(int64)
	return userID, ok
}
