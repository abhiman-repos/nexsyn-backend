package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	// jwtSecret is the secret key used for signing JWT tokens
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	
	// Common errors for better error handling
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrMissingSecret = errors.New("JWT_SECRET environment variable is not set")
)

// TokenConfig holds configuration for JWT tokens
type TokenConfig struct {
	ExpirationHours int
	Issuer          string
}

// DefaultTokenConfig returns the default token configuration
func DefaultTokenConfig() TokenConfig {
	return TokenConfig{
		ExpirationHours: 24,
		Issuer:          "nexsyn-backend",
	}
}

// Claims represents the custom JWT claims
type Claims struct {
	UserID string  `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateTokens generates new JWT tokens for the given user ID
func GenerateTokens(userID string) (string, error) {
	if len(jwtSecret) == 0 {
		return "", ErrMissingSecret
	}

	claims := Claims{
		UserID: userID, // ✅ now matches type
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "nexsyn-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
// GenerateTokenWithExpiration generates a token with custom expiration
func GenerateTokenWithExpiration(userID string, expirationHours int) (string, error) {
	if len(jwtSecret) == 0 {
		return "", ErrMissingSecret
	}
	
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "nexsyn-backend",
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// VerifyToken verifies and parses a JWT token, returning the user ID
func VerifyToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	return claims.UserID, nil // ✅ string
}

// RefreshToken generates a new token from an existing valid token
func RefreshToken(tokenString string) (string, error) {
	userID, err := VerifyToken(tokenString)
	if err != nil {
		return "", err
	}
	
	return GenerateTokens(userID)
}

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", errors.New("failed to hash password")
	}
	
	return string(hashedPassword), nil
}

// CheckPassword compares a hashed password with a plain text password
func CheckPassword(hashedPassword, password string) error {
	if hashedPassword == "" || password == "" {
		return errors.New("password or hash cannot be empty")
	}
	
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)
	
	if err != nil {
		return errors.New("invalid password")
	}
	
	return nil
}

// ValidatePasswordStrength checks if password meets security requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case char >= '!' && char <= '/', char >= ':' && char <= '@', char >= '[' && char <= '`', char >= '{' && char <= '~':
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	
	return nil
}

// GetTokenExpiration returns the expiration time of a token
func GetTokenExpiration(tokenString string) (time.Time, error) {
	if len(jwtSecret) == 0 {
		return time.Time{}, ErrMissingSecret
	}
	
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	
	if err != nil {
		return time.Time{}, err
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return time.Time{}, errors.New("invalid claims")
	}
	
	return claims.ExpiresAt.Time, nil
}