package main

import (
	"os"
	"fmt"
	"log"
	"database/sql"
	"strings"
	"time"
	"net/http"
	"sync/atomic"
	"encoding/json"
	"github.com/andrei-himself/chirpy/internal/database"
	"github.com/andrei-himself/chirpy/internal/auth"
	"github.com/joho/godotenv"   
	"github.com/google/uuid"
)
import _ "github.com/lib/pq"

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
}

type User struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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
	w.Header().Set("Content-Type", "application/json")
	if cfg.platform != "dev" {
		w.WriteHeader(403)
	}
	_ = cfg.fileserverHits.Swap(0)
	err := cfg.db.DeleteUsers(req.Context())
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	err = cfg.db.DeleteChirps(req.Context())
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
}

func (cfg *apiConfig) handleChirps(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	type errResp struct {
		Error string `json:"error"`
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
	createChirpParams := database.CreateChirpParams{
		Body : replaced,
		UserID : params.UserID,
	}
	chirp, err := cfg.db.CreateChirp(req.Context(), createChirpParams)
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
	mapped := Chirp{
		ID : chirp.ID,
		CreatedAt : chirp.CreatedAt,
		UpdatedAt : chirp.UpdatedAt,
		Body : chirp.Body,
		UserID : chirp.UserID,
	}

	dat, err := json.Marshal(mapped)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(201)
	w.Write(dat)
	return
}

func (cfg *apiConfig) handleUsers(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	type errResp struct {
		Error string `json:"error"`
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

	hashed, err := auth.HashPassword(params.Password)
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
	createUserParams := database.CreateUserParams{
		Email : params.Email,
		HashedPassword : sql.NullString{
			String : hashed,
			Valid : true,
		},
	}
	user, err := cfg.db.CreateUser(req.Context(), createUserParams)
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

	mapped := User{
		ID : user.ID,
		CreatedAt : user.CreatedAt,
		UpdatedAt : user.UpdatedAt,
		Email : user.Email,
	}
	dat, err := json.Marshal(mapped)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(201)
	w.Write(dat)
	return
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, req *http.Request) {
	type errResp struct {
		Error string `json:"error"`
	}
	w.Header().Set("Content-Type", "application/json")
	chirps, err := cfg.db.GetChirps(req.Context())
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

	mappedChirps := []Chirp{}
	for _, v := range chirps {
		mapped := Chirp{
			ID : v.ID,
			CreatedAt : v.CreatedAt,
			UpdatedAt : v.UpdatedAt,
			Body : v.Body,
			UserID : v.UserID,
		}
		mappedChirps = append(mappedChirps, mapped)
	}

	dat, err := json.Marshal(mappedChirps)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
	return
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, req *http.Request) {
	type errResp struct {
		Error string `json:"error"`
	}
	chirpID := req.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpID)
	chirp, err := cfg.db.GetChirp(req.Context(), chirpUUID)
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
		w.WriteHeader(404)
		w.Write(dat)
		return
	}

	mapped := Chirp{
		ID : chirp.ID,
		CreatedAt : chirp.CreatedAt,
		UpdatedAt : chirp.UpdatedAt,
		Body : chirp.Body,
		UserID : chirp.UserID,
	}
	dat, err := json.Marshal(mapped)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
	return
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}
	type errResp struct {
		Error string `json:"error"`
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

	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
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
		w.WriteHeader(401)
		w.Write(dat)
		return
	}

	pwMatch, err := auth.CheckPasswordHash(params.Password, user.HashedPassword.String)
	if err != nil || pwMatch == false {
		respBody := errResp{
			Error : "Something went wrong",
		}
		dat, err2 := json.Marshal(respBody)
		if err2 != nil {
			log.Printf("Error marshalling JSON: %s", err2)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(401)
		w.Write(dat)
		return
	}

	mapped := User{
		ID : user.ID,
		CreatedAt : user.CreatedAt,
		UpdatedAt : user.UpdatedAt,
		Email : user.Email,
	}
	dat, err := json.Marshal(mapped)
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
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	apiCfg.db = dbQueries
	apiCfg.platform = platform

	serveMux := http.NewServeMux()
	server := http.Server{
		Addr : ":8080",
		Handler : serveMux,
	}

	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serveMux.HandleFunc("GET /api/healthz", handleHealthz)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handleReset)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.handleChirps)
	serveMux.HandleFunc("POST /api/users", apiCfg.handleUsers)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handleGetChirp)
	serveMux.HandleFunc("POST /api/login", apiCfg.handleLogin)

	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}