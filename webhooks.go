package main

import (
	"chirpy/internal/auth"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		respondWithError(w, 401, err)
		return
	}
	if apiKey != cfg.polkaKey {
		respondWithError(w, 401, err)
		return
	}

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	userUUID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, 400, err)
		return
	}

	err = cfg.db.UpgradeUserByID(r.Context(), userUUID)
	if err != nil {
		respondWithError(w, 404, err)
		return
	}

	w.WriteHeader(204)
}
