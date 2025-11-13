package main

import (
	"fmt"
	"net/http"
)

func handleHealthz(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	_, _ = w.Write([]byte("OK"))
}

func main () {
	serveMux := http.NewServeMux()
	server := http.Server{
		Addr : ":8080",
		Handler : serveMux,
	}

	// serveMux.Handle("/app/", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serveMux.HandleFunc("/healthz", handleHealthz)

	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}

}