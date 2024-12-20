package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/mikarwacki/chirpy/internal/auth"
	"github.com/mikarwacki/chirpy/internal/database"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "Error reading response body", err)
		return
	}

	u := requestUser{}
	err = json.Unmarshal(data, &u)
	if err != nil {
		respondWithError(w, 400, "Error unmarshaling json", err)
		return
	}

	hashedPassword, err := auth.HashPassword(u.Password)
	if err != nil {
		log.Printf("Error hashing password: %v\n", err)
		respondWithError(w, 400, "Error hashing password", err)
		return
	}

	createUserParams := database.CreateUserParams{Email: u.Email, HashedPassword: hashedPassword}
	us, err := cfg.db.CreateUser(r.Context(), createUserParams)
	if err != nil {
		log.Printf("Failed creating user: %v\n", err)
		respondWithError(w, 400, "Error reading from db", err)
		return
	}

	rUser := responseUser{ID: us.ID, CreatedAt: us.CreatedAt, UpdatedAt: us.UpdatedAt, Email: us.Email}
	respondWithJson(w, 201, rUser)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "Error reading response body", err)
		return
	}

	rqUser := requestUser{}
	err = json.Unmarshal(data, &rqUser)
	if err != nil {
		respondWithError(w, 400, "Error unmarshaling json", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), rqUser.Email)
	if err != nil {
		log.Printf("Error fetching: %v", err)
		respondWithError(w, 400, "Error fetching database", err)
		return
	}

	err = auth.CheckPasswordHash(rqUser.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password", err)
		return
	}

	refresh := time.Hour
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, refresh)
	if err != nil {
		respondWithError(w, 400, "Error making jwt token", err)
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 400, "Error creating refresh token", err)
		return
	}
	newToken := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	}
	_, err = cfg.db.CreateRefreshToken(r.Context(), newToken)
	if err != nil {
		respondWithError(w, 400, "Error saving refresh token", err)
		return
	}

	rsUser := *NewResponseUser(user, token, refreshToken)
	respondWithJson(w, 200, rsUser)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	log.Printf("TOKEN: %v\n", token)
	if err != nil {
		respondWithError(w, 400, "Token not present in a header", err)
		return
	}
	dbToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("Database error: %v", err)
		respondWithError(w, 401, "Unautharized", err)
		return
	}

	if dbToken.ExpiresAt.Before(time.Now()) || dbToken.RevokedAt.Valid {
		log.Println("Token is expired")
		respondWithError(w, 401, "Unautharized", err)
		return
	}
	newToken, err := auth.MakeJWT(dbToken.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, 400, "Error refreshing access token", err)
		return
	}
	respondWithJson(w, 200, map[string]string{"token": newToken})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 400, "Bearer token missing from the header", err)
		return
	}
	err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 400, "Error revoking token", err)
		return
	}
	respondWithJson(w, 204, nil)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {

}
