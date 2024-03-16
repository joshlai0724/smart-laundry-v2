package iot

import (
	"sync"

	"github.com/google/uuid"
)

type EdgeMapService struct {
	edges map[uuid.UUID]*Edge
	m     sync.Mutex
}

func NewEdgeMapService() *EdgeMapService {
	return &EdgeMapService{
		edges: make(map[uuid.UUID]*Edge),
	}
}

func (s *EdgeMapService) Add(storeID uuid.UUID, e *Edge) {
	s.m.Lock()
	defer s.m.Unlock()
	s.edges[storeID] = e
}

func (s *EdgeMapService) Delete(storeID uuid.UUID) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.edges, storeID)
}

func (s *EdgeMapService) Get(storeID uuid.UUID) *Edge {
	s.m.Lock()
	defer s.m.Unlock()
	return s.edges[storeID]
}
