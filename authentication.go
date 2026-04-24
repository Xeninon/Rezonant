package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Xeninon/Rezonant/internal/auth"
	"github.com/Xeninon/Rezonant/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, 500, "error decoding json")
		return
	}

	user, err := cfg.db.SelectUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 400, "email not registered")
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Error checking password: %s", err)
		respondWithError(w, 500, "error checking password")
		return
	}

	if !match {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		log.Printf("Error making JWT: %s", err)
		respondWithError(w, 500, "error generating JWT")
		return
	}

	refreshToken := auth.MakeRefreshToken()
	_, err = cfg.db.CreateRefreshToken(
		r.Context(),
		database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    user.ID,
			ExpiresAt: time.Now().UTC().AddDate(0, 0, 60),
		})

	if err != nil {
		log.Printf("Error making refresh token: %s", err)
		respondWithError(w, 500, "error making refresh token")
		return
	}

	respondWithJSON(w, 200, User{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email, IsChirpyRed: user.IsChirpyRed, Token: token, RefreshToken: refreshToken})
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	refreshToken, err := cfg.db.SelectRefreshToken(r.Context(), headerToken)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	if refreshToken.ExpiresAt.Before(time.Now().UTC()) || refreshToken.RevokedAt.Valid {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	jwt, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, time.Hour)
	if err != nil {
		log.Printf("Error making JWT: %s", err)
		respondWithError(w, 500, "error generating JWT")
		return
	}

	respondWithJSON(w, 200, struct {
		Token string `json:"token"`
	}{Token: jwt})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	if err = cfg.db.RevokeRefreshToken(r.Context(), headerToken); err != nil {
		log.Printf("Error revoking refresh token: %s", err)
		respondWithError(w, 500, "error revoking token")
		return
	}

	w.WriteHeader(204)
}
