package idempotency

import (
	"sync"
	"time"
)

// Store almacena IDs de mensajes ya procesados
type Store struct {
	mu       sync.RWMutex
	items    map[string]time.Time
	ttl      time.Duration
}

// NewStore crea un nuevo store con TTL
func NewStore(ttl time.Duration) *Store {
	s := &Store{
		items: make(map[string]time.Time),
		ttl:   ttl,
	}
	go s.cleanup()
	return s
}

// IsProcessed verifica si un message_id ya fue procesado
func (s *Store) IsProcessed(messageID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.items[messageID]
	return exists
}

// MarkProcessed marca un message_id como procesado
func (s *Store) MarkProcessed(messageID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[messageID] = time.Now()
}

// cleanup elimina entradas expiradas
func (s *Store) cleanup() {
	ticker := time.NewTicker(s.ttl / 2)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, ts := range s.items {
			if now.Sub(ts) > s.ttl {
				delete(s.items, id)
			}
		}
		s.mu.Unlock()
	}
}