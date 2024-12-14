package main

import (
	"encoding/json"
	"fmt"
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
		respondWithError(w, 500, "Error reading response body")
		return
	}

	u := requestUser{}
	err = json.Unmarshal(data, &u)
	if err != nil {
		respondWithError(w, 400, "Error unmarshaling json")
		return
	}

	hashedPassword, err := auth.HashPassword(u.Password)
	if err != nil {
		log.Printf("Error hashing password: %v\n", err)
		respondWithError(w, 400, "Error hashing password")
		return
	}

	createUserParams := database.CreateUserParams{Email: u.Email, HashedPassword: hashedPassword}
	us, err := cfg.db.CreateUser(r.Context(), createUserParams)
	if err != nil {
		log.Printf("Failed creating user: %v\n", err)
		respondWithError(w, 400, "Error reading from db")
		return
	}

	rUser := responseUser{ID: us.ID, CreatedAt: us.CreatedAt, UpdatedAt: us.UpdatedAt, Email: us.Email}
	respondWithJson(w, 201, rUser)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "Error reading response body")
		return
	}

	rqUser := requestUser{}
	err = json.Unmarshal(data, &rqUser)
	if err != nil {
		respondWithError(w, 400, "Error unmarshaling json")
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), rqUser.Email)
	if err != nil {
		log.Printf("Error fetching: %v", err)
		respondWithError(w, 400, "Error fetching database")
		return
	}

	err = auth.CheckPasswordHash(rqUser.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	durationStr := "1h"
	if rqUser.TokenExpiryDurationInSeconds != 0 {
		durationStr = fmt.Sprintf("%vs", rqUser.TokenExpiryDurationInSeconds)
	}

	refresh, err := time.ParseDuration(durationStr)
	if err != nil {
		respondWithError(w, 400, "Invalid user expiry duration")
		return
	}
	if refresh.Seconds() > time.Hour.Seconds() {
		refresh = time.Hour
	}
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, refresh)
	if err != nil {
		respondWithError(w, 400, "Error making jwt token")
	}
	rsUser := *NewResponseUser(user, token)

	respondWithJson(w, 200, rsUser)
}
