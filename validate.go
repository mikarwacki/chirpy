package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func middlewareValidate(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const maxChirpLen = 140

		type chirpRequest struct {
			Body   string    `json:"body"`
			UserId uuid.UUID `json:"user_id"`
		}

		defer r.Body.Close()
		dat, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, 500, "Something went wrong")
			return
		}

		chp := chirpRequest{}
		err = json.Unmarshal(dat, &chp)
		if err != nil {
			respondWithError(w, 400, "Something went wrong")
			return
		}

		if len(chp.Body) > maxChirpLen {
			respondWithError(w, 400, "Chrip is too long")
			return
		}

		cleaned := cleanBodyFromProf(chp.Body)
		chp.Body = cleaned
		newBody, err := json.Marshal(chp)
		if err != nil {
			log.Printf("Failed marshaling new body %v", err)
			respondWithError(w, 400, "Failed marshaling new body")
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(newBody))
		next.ServeHTTP(w, r)
	})
}

func cleanBodyFromProf(s string) string {
	illegalWords := map[string]struct{}{"kerfuffle": {}, "sharbert": {}, "fornax": {}}
	words := strings.Split(s, " ")
	for i, word := range words {
		if _, ok := illegalWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}
