package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apphttp "github.com/camilleC/tracker/internal/http"
	"github.com/camilleC/tracker/internal/store"
)

// --- Mock Store ---

type mockStore struct {
	data    map[string]store.PainEntry
	failGet bool
	failSet bool
}

func newMockStore() *mockStore {
	return &mockStore{data: make(map[string]store.PainEntry)}
}

func (m *mockStore) Get(id string) (store.PainEntry, bool) {
	if m.failGet {
		return store.PainEntry{}, false
	}
	e, ok := m.data[id]
	return e, ok
}

func (m *mockStore) Set(id string, entry store.PainEntry) {
	if m.failSet {
		return
	}
	m.data[id] = entry
}

func (m *mockStore) Delete(id string) {
	delete(m.data, id)
}

func (m *mockStore) List() []store.PainEntry {
	entries := make([]store.PainEntry, 0, len(m.data))
	for _, v := range m.data {
		entries = append(entries, v)
	}
	return entries
}

// --- HealthCheck ---

func TestHealthCheck(t *testing.T) {
	h := apphttp.NewHandler(newMockStore())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.HealthCheck(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected application/json content type")
	}
}

// --- GetPain ---

func TestGetPain_Found(t *testing.T) {
	s := newMockStore()
	s.data["123"] = store.PainEntry{
		ID:        "123",
		Timestamp: time.Now().Add(-time.Hour),
		Level:     5,
		Location:  store.Back,
	}

	h := apphttp.NewHandler(s)

	req := httptest.NewRequest(http.MethodGet, "/pain/123", nil)
	req.SetPathValue("id", "123")
	rec := httptest.NewRecorder()

	h.GetPain(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body store.PainEntry
	json.NewDecoder(rec.Body).Decode(&body)
	if body.ID != "123" {
		t.Errorf("expected ID 123, got %s", body.ID)
	}
}

func TestGetPain_NotFound(t *testing.T) {
	h := apphttp.NewHandler(&mockStore{failGet: true, data: make(map[string]store.PainEntry)})

	req := httptest.NewRequest(http.MethodGet, "/pain/123", nil)
	req.SetPathValue("id", "123")
	rec := httptest.NewRecorder()

	h.GetPain(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] == "" {
		t.Errorf("expected error message in body")
	}
}

// --- CreatePain ---

func TestCreatePain_Valid(t *testing.T) {
	h := apphttp.NewHandler(newMockStore())

	entry := store.PainEntry{
		Level:    5,
		Location: store.Back,
	}
	body, _ := json.Marshal(entry)

	req := httptest.NewRequest(http.MethodPost, "/pain", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.CreatePain(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var response store.PainEntry
	json.NewDecoder(rec.Body).Decode(&response)
	if response.ID == "" {
		t.Errorf("expected server-generated ID in response")
	}
	if response.Location != store.Back {
		t.Errorf("expected location back, got %s", response.Location)
	}
}

func TestCreatePain_InvalidJSON(t *testing.T) {
	h := apphttp.NewHandler(newMockStore())

	req := httptest.NewRequest(http.MethodPost, "/pain", bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()

	h.CreatePain(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCreatePain_InvalidLevel(t *testing.T) {
	h := apphttp.NewHandler(newMockStore())

	entry := store.PainEntry{
		Level:    99, // invalid
		Location: store.Back,
	}
	body, _ := json.Marshal(entry)

	req := httptest.NewRequest(http.MethodPost, "/pain", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.CreatePain(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCreatePain_InvalidLocation(t *testing.T) {
	h := apphttp.NewHandler(newMockStore())

	entry := store.PainEntry{
		Level:    5,
		Location: "elbow", // invalid
	}
	body, _ := json.Marshal(entry)

	req := httptest.NewRequest(http.MethodPost, "/pain", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.CreatePain(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- ListPain ---

func TestListPain_Empty(t *testing.T) {
	h := apphttp.NewHandler(newMockStore())

	req := httptest.NewRequest(http.MethodGet, "/pain", nil)
	rec := httptest.NewRecorder()

	h.ListPain(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body []store.PainEntry
	json.NewDecoder(rec.Body).Decode(&body)
	if len(body) != 0 {
		t.Errorf("expected empty list, got %d entries", len(body))
	}
}

func TestListPain_WithEntries(t *testing.T) {
	s := newMockStore()
	s.data["1"] = store.PainEntry{ID: "1", Level: 3, Location: store.Back, Timestamp: time.Now().Add(-time.Hour)}
	s.data["2"] = store.PainEntry{ID: "2", Level: 7, Location: store.Neck, Timestamp: time.Now().Add(-time.Hour)}

	h := apphttp.NewHandler(s)

	req := httptest.NewRequest(http.MethodGet, "/pain", nil)
	rec := httptest.NewRecorder()

	h.ListPain(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body []store.PainEntry
	json.NewDecoder(rec.Body).Decode(&body)
	if len(body) != 2 {
		t.Errorf("expected 2 entries, got %d", len(body))
	}
}

// --- UpdatePain ---

func TestUpdatePain_Valid(t *testing.T) {
	s := newMockStore()
	s.data["123"] = store.PainEntry{ID: "123", Level: 3, Location: store.Back, Timestamp: time.Now().Add(-time.Hour)}

	h := apphttp.NewHandler(s)

	updated := store.PainEntry{Level: 8, Location: store.Knee}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/pain/123", bytes.NewBuffer(body))
	req.SetPathValue("id", "123")
	rec := httptest.NewRecorder()

	h.UpdatePain(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var response store.PainEntry
	json.NewDecoder(rec.Body).Decode(&response)
	if response.Level != 8 {
		t.Errorf("expected level 8, got %d", response.Level)
	}
	if response.ID != "123" {
		t.Errorf("expected ID to be preserved as 123, got %s", response.ID)
	}
}

func TestUpdatePain_NotFound(t *testing.T) {
	h := apphttp.NewHandler(newMockStore())

	updated := store.PainEntry{Level: 8, Location: store.Knee}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/pain/123", bytes.NewBuffer(body))
	req.SetPathValue("id", "123")
	rec := httptest.NewRecorder()

	h.UpdatePain(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestUpdatePain_InvalidJSON(t *testing.T) {
	s := newMockStore()
	s.data["123"] = store.PainEntry{ID: "123", Level: 3, Location: store.Back, Timestamp: time.Now().Add(-time.Hour)}

	h := apphttp.NewHandler(s)

	req := httptest.NewRequest(http.MethodPut, "/pain/123", bytes.NewBufferString("not json"))
	req.SetPathValue("id", "123")
	rec := httptest.NewRecorder()

	h.UpdatePain(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestUpdatePain_InvalidLevel(t *testing.T) {
	s := newMockStore()
	s.data["123"] = store.PainEntry{ID: "123", Level: 3, Location: store.Back, Timestamp: time.Now().Add(-time.Hour)}

	h := apphttp.NewHandler(s)

	updated := store.PainEntry{Level: 99, Location: store.Knee}
	body, _ := json.Marshal(updated)

	req := httptest.NewRequest(http.MethodPut, "/pain/123", bytes.NewBuffer(body))
	req.SetPathValue("id", "123")
	rec := httptest.NewRecorder()

	h.UpdatePain(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- DeletePain ---

func TestDeletePain_Valid(t *testing.T) {
	s := newMockStore()
	s.data["123"] = store.PainEntry{ID: "123", Level: 3, Location: store.Back, Timestamp: time.Now().Add(-time.Hour)}

	h := apphttp.NewHandler(s)

	req := httptest.NewRequest(http.MethodDelete, "/pain/123", nil)
	req.SetPathValue("id", "123")
	rec := httptest.NewRecorder()

	h.DeletePain(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}

	if _, ok := s.data["123"]; ok {
		t.Errorf("expected entry to be deleted from store")
	}
}

func TestDeletePain_NotFound(t *testing.T) {
	h := apphttp.NewHandler(newMockStore())

	req := httptest.NewRequest(http.MethodDelete, "/pain/123", nil)
	req.SetPathValue("id", "123")
	rec := httptest.NewRecorder()

	h.DeletePain(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}