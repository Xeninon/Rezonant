package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Xeninon/Rezonant/internal/auth"
	"github.com/Xeninon/Rezonant/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func profaneFilter(msg string) string {
	words := strings.Split(msg, " ")
	for i, word := range words {
		word = strings.ToLower(word)
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "error decoding json")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "expired JWT")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: profaneFilter(params.Body), UserID: userID})
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		respondWithError(w, 500, "error creating chirp")
		return

	}

	respondWithJSON(w, 201, Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID})
}

func (cfg *apiConfig) handlerReadChirps(w http.ResponseWriter, r *http.Request) {
	stringID := r.URL.Query().Get("author_id")
	authorID, err := uuid.Parse(stringID)
	if stringID == "" || err != nil {
		chirps, err := cfg.db.SelectChirps(r.Context())
		if err != nil {
			log.Printf("Error reading chirps: %s", err)
			respondWithError(w, 500, "error reading chirp")
			return
		}

		returnVals := make([]Chirp, len(chirps))
		for i, chirp := range chirps {
			returnVals[i] = Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID}
		}

		respondWithJSON(w, 200, returnVals)
	}

	chirps, err := cfg.db.SelectChirpsByUser(r.Context(), authorID)
	if err != nil {
		log.Printf("Error reading chirps: %s", err)
		respondWithError(w, 500, "error reading chirp")
		return
	}

	returnVals := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		returnVals[i] = Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.UserID}
	}

	respondWithJSON(w, 200, returnVals)
}

func (cfg *apiConfig) handlerReadChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	chirp, err := cfg.db.SelectChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	respondWithJSON(w, 200, Chirp{ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, Body: chirp.Body, UserID: chirp.ID})
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	headderToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(headderToken, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "expired JWT")
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	chirp, err := cfg.db.SelectChirp(r.Context(), chirpID)
	if err != nil {
		log.Printf("Error getting chirp: %s", err)
		respondWithError(w, 404, "chrip not found")
		return
	}

	if userID != chirp.UserID {
		respondWithError(w, 403, "Unauthorized")
		return
	}

	if err := cfg.db.DeleteChirp(r.Context(), chirpID); err != nil {
		log.Printf("Error deleting chirp: %s", err)
		respondWithError(w, 500, "error deleting chirp")
		return
	}

	w.WriteHeader(204)
}
