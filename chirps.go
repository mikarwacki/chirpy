package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/mikarwacki/chirpy/internal/database"
)

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	userId := req.Context().Value("userId").(uuid.UUID)
	type chirp struct {
		Body string `json:"body"`
	}

	defer req.Body.Close()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		respondWithError(w, 500, "Error reading request body", err)
	}
	log.Println(string(data))

	chir := chirp{}
	err = json.Unmarshal(data, &chir)
	if err != nil {
		respondWithError(w, 400, "Error unmarshalling data", err)
		return
	}

	dbChirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{Body: chir.Body, UserID: userId})
	if err != nil {
		respondWithError(w, 400, "Error creating chirp", err)
	}

	rChirp := responseChirp{ID: dbChirp.ID, CreatedAt: dbChirp.CreatedAt, UpdatedAt: dbChirp.UpdatedAt, Body: dbChirp.Body, UserID: dbChirp.UserID}
	respondWithJson(w, 201, rChirp)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")
	sortStrat := r.URL.Query().Get("sort")

	var chirps []database.Chirp
	var err error
	if authorId == "" {
		chirps, err = cfg.db.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, 400, "Error getting chirps", err)
			return
		}
	} else {
		userUuid, err := uuid.Parse(authorId)
		if err != nil {
			respondWithError(w, 400, "Error parsing author uuid", err)
			return
		}

		chirps, err = cfg.db.GetChirpsByAuthor(r.Context(), userUuid)
		if err != nil {
			respondWithError(w, 400, "Error getting chirps", err)
			return
		}
	}

	rChirps := make([]responseChirp, len(chirps))
	for i, chr := range chirps {
		rChirps[i] = responseChirp{ID: chr.ID, CreatedAt: chr.CreatedAt, UpdatedAt: chr.UpdatedAt, Body: chr.Body, UserID: chr.UserID}
	}

	sort.Slice(rChirps, func(i, j int) bool {
		if sortStrat == "desc" {
			return rChirps[i].CreatedAt.After(rChirps[j].CreatedAt)
		}
		return rChirps[i].CreatedAt.Before(rChirps[j].CreatedAt)
	})
	respondWithJson(w, 200, rChirps)
}

func (cfg *apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {
	chirpId, err := uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		respondWithError(w, 400, "Invalid chirp id", err)
	}
	chirp, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 404, "Chirp doesn't exist", err)
	}

	rChirp := responseChirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID}
	respondWithJson(w, 200, rChirp)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userId").(uuid.UUID)
	chirpId, err := uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		respondWithError(w, 400, "Error parsing uuid", err)
		return
	}
	dbChirp, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 400, "Chirp doesn't exist", err)
		return
	}

	if dbChirp.UserID != userId {
		respondWithError(w, 403, "Current user isn't author of the chirp", err)
		return
	}
	err = cfg.db.DeleteChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 404, "Error deleting the chirp, chirp doesn't exist", err)
		return
	}

	respondWithJson(w, 204, nil)
}
