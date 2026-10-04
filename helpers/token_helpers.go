package helpers

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/cd-Ishita/nutriediet-go/database"
	jwt "github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	UserID    string
	UserType  string
	jwt.RegisteredClaims
}

var SECRET_KEY string
var jwtInitialized = false

// initJWTSecret initializes and validates the JWT secret key
// Called on first use instead of in init() to ensure env vars are loaded first
func initJWTSecret() {
	if jwtInitialized {
		return
	}

	SECRET_KEY = os.Getenv("JWT_SECRET_KEY")

	// Validate JWT secret key exists
	if SECRET_KEY == "" {
		log.Fatal("❌ JWT_SECRET_KEY environment variable is required. Generate one with: openssl rand -base64 64")
	}

	// Validate minimum length (256 bits = 32 bytes minimum)
	if len(SECRET_KEY) < 32 {
		log.Fatal("❌ JWT_SECRET_KEY must be at least 32 characters long for security")
	}

	jwtInitialized = true
	log.Println("✅ JWT Secret Key loaded successfully")
}

func GenerateAllTokens(email, firstName, lastName, userType string, id uint64) (string, string, error) {
	// Initialize JWT secret on first use
	initJWTSecret()

	accessTTL := 15 * time.Minute
	refreshTTL := 90 * 24 * time.Hour
	if userType == "ADMIN" {
		// Waiting-room display stays logged in for clinic hours; avoid 15-minute dropouts.
		accessTTL = 180 * 24 * time.Hour
		refreshTTL = 180 * 24 * time.Hour
	}

	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UserType:  userType,
		UserID:    strconv.FormatUint(id, 10),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshClaims := &SignedDetails{
		Email:  email,
		UserID: strconv.FormatUint(id, 10),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %v", err)
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %v", err)
	}

	return token, refreshToken, nil
}

func UpdateTokens(token, refreshToken string, id uint64) error {
	db := database.DB
	err := db.Table("user_auths").Where("id = ?", id).Updates(map[string]interface{}{
		"token":         token,
		"refresh_token": refreshToken,
	}).Error
	if err != nil {
		fmt.Println("error: cannot update the tokens")
		return err
	}
	return nil
}

func ValidateToken(token string) (SignedDetails, error) {
	// Initialize JWT secret on first use
	initJWTSecret()

	res, err := jwt.ParseWithClaims(token, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return SignedDetails{}, fmt.Errorf("error parsing token: %v", err)
	}

	claims, ok := res.Claims.(*SignedDetails)
	if !ok {
		return SignedDetails{}, errors.New("invalid token claims")
	}

	if claims == nil {
		return SignedDetails{}, errors.New("token claims are nil")
	}

	// Check if token is expired
	if claims.ExpiresAt.Before(time.Now()) {
		return SignedDetails{}, errors.New("token has expired")
	}

	return *claims, nil
}
