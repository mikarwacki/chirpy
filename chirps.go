package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/mikarwacki/chirpy/internal/database"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	type chirp struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	defer req.Body.Close()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		respondWithError(w, 500, "Error reading request body")
	}
	log.Println(string(data))

	chir := chirp{}
	err = json.Unmarshal(data, &chir)
	if err != nil {
		log.Printf("error unmarshalling data: %v", err)
		respondWithError(w, 400, "Error unmarshalling data")
		return
	}

	log.Println(chir.UserId)
	dbChirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{Body: chir.Body, UserID: chir.UserId})
	if err != nil {
		log.Printf("Error creating chirp: %v", err)
		respondWithError(w, 400, "Error creating chirp")
	}

	rChirp := responseChirp{ID: dbChirp.ID, CreatedAt: dbChirp.CreatedAt, UpdatedAt: dbChirp.UpdatedAt, Body: dbChirp.Body, UserID: dbChirp.UserID}
	respondWithJson(w, 201, rChirp)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, 400, "Error getting chirps")
		return
	}
	rChirps := make([]responseChirp, len(chirps))
	for i, chr := range chirps {
		rChirps[i] = responseChirp{ID: chr.ID, CreatedAt: chr.CreatedAt, UpdatedAt: chr.UpdatedAt, Body: chr.Body, UserID: chr.UserID}
	}

	respondWithJson(w, 200, rChirps)
}

func (cfg *apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {
	chirpId, err := uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		log.Printf("Invalid uuid %v", err)
		respondWithError(w, 400, "Invalid chirp id")
	}
	chirp, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 404, "Chirp doesn't exist")
	}

	rChirp := responseChirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID}
	respondWithJson(w, 200, rChirp)
}
