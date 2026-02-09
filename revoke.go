package main

import (
	"chirpy/internal/auth"
	"net/http"
)

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	err = cfg.db.RevokeToken(r.Context(), bearer)
	if err != nil {
		respondWithError(w, 400, err)
	}

	w.WriteHeader(204)
}
