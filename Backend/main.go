package main

import (
	"log"
	"net/http"

	"CasinoLobbyBE/db"
	"CasinoLobbyBE/routes"
)

func main() {
	db.Connect()
	db.InitSchemas()

	mux := http.NewServeMux()

	routes.RegisterRoutes(mux)

	log.Println("🚀 Server running on :8080")
	http.ListenAndServe(":8080", mux)
}
