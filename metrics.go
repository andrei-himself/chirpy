package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	val := cfg.fileserverHits.Load()
	strVal := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
	`, val)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write([]byte(strVal))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
