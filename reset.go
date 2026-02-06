package main

import (
	"errors"
	"net/http"
)

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, 403, errors.New("Reset can only be performed on a local dev environment!"))
		return
	}

	cfg.fileserverHits.Store(0)
	err := cfg.db.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, err)
		return
	}

	body := "Reset metrics count to 0. Reset users database."
	w.Header().Set("Content-Type", "text/plain: charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(body))
}
