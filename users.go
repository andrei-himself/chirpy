package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {
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

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 400, err)
	}

	userParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
	}

	user, err := cfg.db.CreateUser(r.Context(), userParams)
	if err != nil {
		respondWithError(w, 500, err)
		return
	}

	respondWithJSON(w, 201, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	})
}

func (cfg *apiConfig) handlePutUsers(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err)
		return
	}

	userUUID, err := auth.ValidateJWT(bearer, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, 401, err)
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	hashed, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	updateUserParams := database.UpdateEmailAndPasswordParams{
		ID:             userUUID,
		Email:          params.Email,
		HashedPassword: hashed,
	}
	user, err := cfg.db.UpdateEmailAndPassword(r.Context(), updateUserParams)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	respondWithJSON(w, 200, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	})
}
