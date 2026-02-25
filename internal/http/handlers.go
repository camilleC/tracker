package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/camilleC/tracker/internal/logger"
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

	mux.HandleFunc("POST /pain", h.CreatePain)
	mux.HandleFunc("GET /pain", h.ListPain)

	mux.HandleFunc("GET /pain/{id}", h.GetPain)
	mux.HandleFunc("PUT /pain/{id}", h.UpdatePain)
	mux.HandleFunc("DELETE /pain/{id}", h.DeletePain)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) GetPain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	entry, ok := h.store.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found") // fix: 404 not 400
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (h *Handler) CreatePain(w http.ResponseWriter, r *http.Request) {
	var entry store.PainEntry

	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	entry.ID = uuid.New().String()
	entry.Timestamp = time.Now()

	if err := entry.Validate(); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error()) // fix: single response, use actual error
		return
	}

	h.store.Set(entry.ID, entry)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

func (h *Handler) UpdatePain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_, ok := h.store.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}

	var entry store.PainEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Preserve server-controlled fields
	entry.ID = id
	entry.Timestamp = time.Now()

	if err := entry.Validate(); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.store.Set(id, entry)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

func (h *Handler) DeletePain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	_, ok := h.store.Get(id)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}

	h.store.Delete(id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListPain(w http.ResponseWriter, r *http.Request) {
	entries := h.store.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	logger.Error(msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}