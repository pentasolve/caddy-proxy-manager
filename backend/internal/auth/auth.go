package auth

import (
	"caddy-proxy-manager/internal/db"
	"caddy-proxy-manager/internal/models"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)
const (
	argonTime    = 3
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLength   = 16
)
const (
	AccessTokenExpiry  = 15 * time.Minute
	RefreshTokenExpiry = 7 * 24 * time.Hour
)

var (
	accessSecretKey  []byte
	refreshSecretKey []byte
	secretKeyOnce    sync.Once
)
var SecretKey []byte

func init() {
	initSecretKeys()
	go cleanupRevokedTokens()
}

func initSecretKeys() {
	secretKeyOnce.Do(func() {
		accessSecret := os.Getenv("JWT_ACCESS_SECRET")
		refreshSecret := os.Getenv("JWT_REFRESH_SECRET")

		if accessSecret == "" {
			accessSecretKey = generateRandomBytes(32)
		} else {
			accessSecretKey = []byte(accessSecret)
		}

		if refreshSecret == "" {
			refreshSecretKey = generateRandomBytes(32)
		} else {
			refreshSecretKey = []byte(refreshSecret)
		}

		SecretKey = accessSecretKey
	})
}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}
type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"type"`
	jwt.RegisteredClaims
}
type RefreshClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"type"`
	TokenID   string `json:"jti"`
	jwt.RegisteredClaims
}
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads, b64Salt, b64Hash), nil
}
func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	var memory, time uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	comparisonHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(hash)))

	return subtle.ConstantTimeCompare(hash, comparisonHash) == 1, nil
}
func LoginWithTokens(username, password string) (*TokenPair, error) {
	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		time.Sleep(100 * time.Millisecond)
		return nil, errors.New("invalid credentials")
	}
	if strings.HasPrefix(user.Password, "$argon2id$") {
		valid, err := VerifyPassword(password, user.Password)
		if err != nil || !valid {
			return nil, errors.New("invalid credentials")
		}
	} else {
		if user.Password != password {
			return nil, errors.New("invalid credentials")
		}
		if hashedPassword, err := HashPassword(password); err == nil {
			user.Password = hashedPassword
			db.DB.Save(&user)
		}
	}

	return generateTokenPair(user.ID, user.Username)
}
func Login(username, password string) (string, error) {
	pair, err := LoginWithTokens(username, password)
	if err != nil {
		return "", err
	}
	return pair.AccessToken, nil
}

func generateTokenPair(userID uint, username string) (*TokenPair, error) {
	now := time.Now()
	tokenID := generateTokenID()
	accessClaims := &Claims{
		UserID:    userID,
		Username:  username,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "caddy-proxy-manager",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(accessSecretKey)
	if err != nil {
		return nil, err
	}
	refreshClaims := &RefreshClaims{
		UserID:    userID,
		Username:  username,
		TokenType: "refresh",
		TokenID:   tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "caddy-proxy-manager",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(refreshSecretKey)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(AccessTokenExpiry.Seconds()),
	}, nil
}

func generateTokenID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
func RefreshTokens(refreshTokenString string) (*TokenPair, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(refreshTokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return refreshSecretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type")
	}

	if isTokenRevoked(claims.TokenID) {
		return nil, errors.New("token has been revoked")
	}

	var user models.User
	if err := db.DB.First(&user, claims.UserID).Error; err != nil {
		return nil, errors.New("user not found")
	}
	revokeToken(claims.TokenID, claims.UserID, claims.ExpiresAt.Time)

	return generateTokenPair(claims.UserID, claims.Username)
}
func ValidateAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return accessSecretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
func Logout(refreshTokenString string) error {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(refreshTokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return refreshSecretKey, nil
	})

	if err != nil || !token.Valid {
		return errors.New("invalid token")
	}

	revokeToken(claims.TokenID, claims.UserID, claims.ExpiresAt.Time)
	return nil
}
func revokeToken(tokenID string, userID uint, expiresAt time.Time) {
	revokedToken := models.RevokedToken{
		TokenID:   tokenID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		RevokedAt: time.Now(),
	}
	db.DB.Where(models.RevokedToken{TokenID: tokenID}).FirstOrCreate(&revokedToken)
}
func isTokenRevoked(tokenID string) bool {
	var count int64
	db.DB.Model(&models.RevokedToken{}).Where("token_id = ?", tokenID).Count(&count)
	return count > 0
}
func cleanupRevokedTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		db.DB.Where("expires_at < ?", time.Now()).Delete(&models.RevokedToken{})
	}
}
func RevokeAllUserTokens(userID uint) error {
	revokedToken := models.RevokedToken{
		TokenID:   fmt.Sprintf("all_user_%d_%d", userID, time.Now().Unix()),
		UserID:    userID,
		ExpiresAt: time.Now().Add(RefreshTokenExpiry),
		RevokedAt: time.Now(),
	}
	return db.DB.Create(&revokedToken).Error
}
func Refresh(tokenString string) (string, error) {
	pair, err := RefreshTokens(tokenString)
	if err != nil {
		return "", err
	}
	return pair.AccessToken, nil
}
