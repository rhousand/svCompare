package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rhousand/svcompare/internal/db"
	"github.com/rhousand/svcompare/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Authenticator is the interface for authentication strategies.
// LocalAuthenticator implements username/password auth.
// Future: implement GoogleAuthenticator for OAuth 2.0.
type Authenticator interface {
	Login(ctx context.Context, username, password string) (*models.User, error)
}

// LocalAuthenticator authenticates users via bcrypt-hashed passwords in SQLite.
type LocalAuthenticator struct {
	db        *db.DB
	jwtSecret string
}

// NewLocalAuthenticator creates a new LocalAuthenticator.
func NewLocalAuthenticator(database *db.DB, jwtSecret string) *LocalAuthenticator {
	return &LocalAuthenticator{db: database, jwtSecret: jwtSecret}
}

// Login verifies credentials and returns the authenticated user.
func (a *LocalAuthenticator) Login(_ context.Context, username, password string) (*models.User, error) {
	row, err := a.db.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("db lookup: %w", err)
	}
	if row == nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(row.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return &models.User{ID: row.ID, Username: row.Username}, nil
}

// HashPassword hashes a plaintext password with bcrypt at default cost.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

// Claims are the JWT payload fields.
type Claims struct {
	jwt.RegisteredClaims
}

// IssueToken creates a signed JWT for a given user ID.
func IssueToken(userID, secret string, expiry time.Duration) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken parses and validates a JWT string, returning the subject (user ID).
func ValidateToken(tokenStr, secret string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}
	return claims.Subject, nil
}
