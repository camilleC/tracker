package http

import (
	"time"
	"encoding/json"
	"net/http"
	"github.com/google/uuid"
	"github.com/camilleC/tracker/internal/store"
)

type Handler struct {
	store store.PainStore
}

func NewHandler(s store.PainStore) *Handler {
	return &Handler{store: s}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.HealthCheck)
	mux.HandleFunc("/pain", h.HandlePain)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) HandlePain(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        h.GetPain(w, r)
    case http.MethodPost:
        h.CreatePain(w, r)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}

func (h *Handler) GetPain(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	entry, ok := h.store.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (h *Handler) CreatePain(w http.ResponseWriter, r *http.Request) {
	var entry store.PainEntry

	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Set server-controlled fields
	entry.ID = uuid.New().String()
	entry.Timestamp = time.Now()

	if err := entry.ValidateLocation(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := entry.ValidateLevel(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.store.Set(entry.ID, entry)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}