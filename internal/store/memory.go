package store

import (
	"errors"
	"sync"

	"github.com/ianovski/ai-change-risk-gate/internal/model"
)

var ErrNotFound = errors.New("approval record not found")

type ApprovalStore interface {
	Put(record model.ApprovalRecord)
	Get(id string) (model.ApprovalRecord, error)
	Update(record model.ApprovalRecord) error
}

type MemoryApprovalStore struct {
	mu   sync.RWMutex
	data map[string]model.ApprovalRecord
}

func NewMemoryApprovalStore() *MemoryApprovalStore {
	return &MemoryApprovalStore{data: make(map[string]model.ApprovalRecord)}
}

func (s *MemoryApprovalStore) Put(record model.ApprovalRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[record.ID] = record
}

func (s *MemoryApprovalStore) Get(id string) (model.ApprovalRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.data[id]
	if !ok {
		return model.ApprovalRecord{}, ErrNotFound
	}
	return rec, nil
}

func (s *MemoryApprovalStore) Update(record model.ApprovalRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[record.ID]; !ok {
		return ErrNotFound
	}
	s.data[record.ID] = record
	return nil
}
