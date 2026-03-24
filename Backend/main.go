package main

import (
	"log"
	"net/http"

	"CasinoLobbyBE/db"
	"CasinoLobbyBE/routes"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	db.Connect()
	db.InitSchemas()

	mux := http.NewServeMux()

	routes.RegisterRoutes(mux)

	log.Println("🚀 Server running on :8080")
	http.ListenAndServe(":8080", enableCORS(mux))
}
