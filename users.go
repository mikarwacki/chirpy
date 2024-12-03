package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type user struct {
		Email string `json:"email"`
	}

	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 500, "Error reading response body")
		return
	}

	u := user{}
	err = json.Unmarshal(data, &u)
	if err != nil {
		respondWithError(w, 400, "Error unmarshaling json")
		return
	}

	us, err := cfg.db.CreateUser(r.Context(), u.Email)
	if err != nil {
		log.Printf("Failed creating user: %v\n", err)
		respondWithError(w, 400, "Error reading from db")
		return
	}

	rUser := responseUser{ID: us.ID, CreatedAt: us.CreatedAt, UpdatedAt: us.UpdatedAt, Email: us.Email}
	respondWithJson(w, 201, rUser)
}
