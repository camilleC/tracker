package store

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type PainLocation string

const (
	Back PainLocation = "back"
	Neck PainLocation = "neck"
	Shoulder PainLocation = "shoulder"
	Knee PainLocation = "knee"
	Ankle PainLocation = "ankle"
)

type PainEntry struct {
    ID        string    `json:"id"`
    Timestamp time.Time `json:"timestamp"`
    Level     int       `json:"level"` // 0 - 10 change to enum later?
    Location  PainLocation    `json:"location"`
    Notes     string    `json:"notes"`     // optional
}

type PainStore interface {
	Get(key string) (PainEntry, bool)
	Set(key string, value PainEntry)
	Delete(key string)
	List() []PainEntry
}

type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]PainEntry
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]PainEntry),
	}
}

func (s *MemoryStore) Get(key string) (PainEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *MemoryStore) Set(key string, value PainEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *MemoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *MemoryStore) List() []PainEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries := make([]PainEntry, 0, len(s.data))
	for _, v := range s.data {
		entries = append(entries, v)
	}
	return entries
}

func (p PainEntry) Validate() error {
	if p.Level < 0 || p.Level > 10 {
		return errors.New("level must be 0-10")
	}

	switch p.Location {
	case Back, Neck, Shoulder, Knee, Ankle:
		// valid
	default:
		return fmt.Errorf("invalid location: %s", p.Location)
	}

	if p.Timestamp.After(time.Now()) {
		return errors.New("timestamp cannot be in the future")
	}

	return nil
}