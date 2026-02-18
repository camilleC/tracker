package main

import (
	"log"
	"net/http"

	handler "github.com/camilleC/tracker/internal/http"
	//"github.com/camilleC/tracker/internal/logger"
	"github.com/camilleC/tracker/internal/store"
)

func main() {

	s := store.NewMemoryStore()
	h := handler.NewHandler(s)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	wrapped := handler.LoggingMiddleware(mux)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", wrapped); err != nil {
		log.Fatal(err)
	}
}
