package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	currTime := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: "chirpy", IssuedAt: &jwt.NumericDate{Time: currTime}, ExpiresAt: &jwt.NumericDate{Time: currTime.Add(expiresIn)}, Subject: userID.String()})

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
	if len(split) < 2 {
		return "", fmt.Errorf("Header doesn't contain bearer token")
	}
	return split[1], nil
}
