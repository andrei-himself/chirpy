package main

import (
	"chirpy/internal/auth"
	"errors"
	"net/http"
	"time"
)

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), bearer)
	if err != nil {
		respondWithError(w, 401, err)
		return
	}

	now := time.Now().UTC()
	if !refreshToken.ExpiresAt.After(now) {
		respondWithError(w, 401, errors.New("Refresh token is expired!"))
		return
	}
	if refreshToken.RevokedAt.Valid {
		respondWithError(w, 401, errors.New("Refresh token is revoked!"))
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), refreshToken.UserID)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	type okResp struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, 200, okResp{
		Token: jwt,
	})
}
