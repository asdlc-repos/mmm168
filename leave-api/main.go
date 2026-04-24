package main

import (
	"log"
	"net/http"
	"os"

	"leave-api/internal/handlers"
	"leave-api/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	s := store.New()
	h := handlers.New(s)

	mux := http.NewServeMux()
	handler := h.RegisterRoutes(mux)

	addr := ":" + port
	log.Printf("leave-api starting on %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
