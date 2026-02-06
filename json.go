package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func respondWithError(w http.ResponseWriter, code int, err error) {
	type errResp struct {
		Error string `json:"error"`
	}
	respBody := errResp{
		Error: fmt.Sprintf("%v", err),
	}
	respondWithJSON(w, code, respBody)
}

func respondWithJSON(w http.ResponseWriter, code int, respBody interface{}) {
	dat, err := json.Marshal(respBody)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}
