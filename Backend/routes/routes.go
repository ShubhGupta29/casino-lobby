package routes

import (
	"log"
	"net/http"

	handlers "CasinoLobbyBE/http/handlers"
)

func RegisterRoutes(mux *http.ServeMux) {
	endpoints := []struct {
		pattern string
		handler http.HandlerFunc
	}{
		{"POST /generic/v1/operator/authenticate/{partner_id}", handlers.Authenticate},
		{"POST /generic/v1/operator/balance/{partner_id}", handlers.Balance},
		{"POST /generic/v1/operator/debit/{partner_id}", handlers.Debit},
		{"POST /generic/v1/operator/credit/{partner_id}", handlers.Credit},
		{"POST /generic/v1/operator/rollback/{partner_id}", handlers.Rollback},
		{"POST /lobby/v1/register", handlers.Register},
		{"POST /lobby/v1/login", handlers.Login},
		{"POST /lobby/v1/logout", handlers.Logout},
	}

	log.Println("📍 Registered Endpoints:")
	for _, e := range endpoints {
		mux.HandleFunc(e.pattern, e.handler)
		log.Printf("  - %s\n", e.pattern)
	}
}
