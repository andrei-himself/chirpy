package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}
	if !match {
		respondWithError(w, 401, errors.New("Wrong password!"))
		return
	}

	expiresIn, err := time.ParseDuration("1h")
	if err != nil {
		respondWithError(w, 400, err)
		return
	}
	jwt, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expiresIn)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	refreshStr, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshStr,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60).UTC(),
		RevokedAt: sql.NullTime{},
	}

	refreshToken, err := cfg.db.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	respondWithJSON(w, 200, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     jwt,
		Refresh:   refreshToken.Token,
	})
}
