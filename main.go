package main

import (
	"fmt"
	"log"
	"strings"
	"net/http"
	"sync/atomic"
	"encoding/json"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handleHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, _ = w.Write([]byte("OK"))
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	_, _ = w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>\n", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, req *http.Request) {
	_ = cfg.fileserverHits.Swap(0)
}

func handleValidate(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type errResp struct {
		Error string `json:"error"`
	}
	type okResp struct {
		CleandedBody string `json:"cleaned_body"`
	}
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respBody := errResp{
			Error : "Something went wrong",
		}
		dat, err2 := json.Marshal(respBody)
		if err2 != nil {
			log.Printf("Error marshalling JSON: %s", err2)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	if len(params.Body) > 140 {
		respBody := errResp{
			Error : "Chirp is too long",
		}
		dat, err2 := json.Marshal(respBody)
		if err2 != nil {
			log.Printf("Error marshalling JSON: %s", err2)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	replaced := censorString(params.Body)
	respBody := okResp{
		CleandedBody : replaced,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
	return
}

func censorString(s string) string {
	censored := []string{}
	splitted := strings.Split(s, " ")
	for _, v := range splitted {
		v = strings.ReplaceAll(v, " ", "")
		switch strings.ToLower(v) {
		case "kerfuffle", "sharbert", "fornax":
			censored = append(censored, "****")
		default:
			censored = append(censored, v)
		}
	}
	return strings.Join(censored, " ")
}

func main () {
	apiCfg := apiConfig{}

	serveMux := http.NewServeMux()
	server := http.Server{
		Addr : ":8080",
		Handler : serveMux,
	}

	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /api/healthz", handleHealthz)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handleReset)
	serveMux.HandleFunc("POST /api/validate_chirp", handleValidate)

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}

}