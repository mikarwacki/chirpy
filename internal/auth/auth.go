package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenType string

const (
	TokenTypeAccess TokenType = "chirpy-access"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), len(password))
	if err != nil {
		log.Printf("Error hashing password: %v\n", err)
		return "", err
	}
	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Printf("Error comparing password: %v\n", err)
		return err
	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})

	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		log.Printf("Error signing JWT %v", err)
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		log.Printf("Error parsing token: %v", err)
		return uuid.UUID{}, err
	}

	id, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Error getting user id: %v", err)
		return uuid.UUID{}, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	userId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Error parsing id string to uuid %v", err)
		return uuid.UUID{}, err
	}
	return userId, nil
}

func GetBearerToken(header http.Header) (string, error) {
	section := header.Get("Authorization")
	split := strings.Split(section, " ")
	if len(split) < 2 || split[0] != "Bearer" {
		return "", fmt.Errorf("Header doesn't contain bearer token")
	}
	return split[1], nil
}

func GetAPIKey(header http.Header) (string, error) {
	section := header.Get("Authorization")
	log.Println(section)
	split := strings.Split(section, " ")
	if len(split) < 2 || split[0] != "ApiKey" {
		return "", fmt.Errorf("Header doesn't contain api key")
	}
	return split[1], nil
}

func MakeRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Printf("Error creating access token: %v", err)
		return "", err
	}
	token := hex.EncodeToString(bytes)
	return token, nil
}
