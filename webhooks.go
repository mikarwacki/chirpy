package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/mikarwacki/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerWebhookPolka(w http.ResponseWriter, r *http.Request) {
	type PolkaRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserId uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	apiKey, err := auth.GetAPIKey(r.Header)
	log.Printf("APIKEY: %v, Error: %v", apiKey, err)
	if err != nil {
		respondWithError(w, 401, "Missing apikey", err)
		return
	}
	if apiKey != cfg.polkaApiKey {
		respondWithError(w, 401, "Unauthorized", nil)
		return
	}

	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 400, "Error reading request body", err)
		return
	}

	polkaRq := PolkaRequest{}
	err = json.Unmarshal(data, &polkaRq)
	if err != nil {
		respondWithError(w, 400, "Error unmarshaling request", err)
		return
	}

	if polkaRq.Event != "user.upgraded" {
		respondWithJson(w, 204, nil)
		return
	}
	log.Printf("UserID in request: %v", polkaRq.Data.UserId)

	err = cfg.db.MarkUserRedById(r.Context(), polkaRq.Data.UserId)
	if err != nil {
		respondWithError(w, 404, "User doesn't exist", err)
		return
	}

	respondWithJson(w, 204, nil)
}
