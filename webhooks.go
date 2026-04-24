package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Xeninon/Rezonant/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type data struct {
		UserID string `json:"user_id"`
	}

	type parameters struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	if apiKey != cfg.polkaKey {
		w.WriteHeader(401)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		log.Printf("Parse: %s", err)
		w.WriteHeader(404)
		return
	}

	if err = cfg.db.UpdateChirpyRed(r.Context(), userID); err != nil {
		log.Printf("Update: %s", err)
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(204)
}
