package mock

import (
	"github.com/ngalayko/dyn-dns/app/provider"
)

// Mock is a mock dns provider.
type Mock struct {
	recordsList []*provider.Record
	recordsMap  map[string]*provider.Record
}

// New is a mock dns constructor.
func New() *Mock {
	return &Mock{
		recordsList: make([]*provider.Record, 0),
		recordsMap:  make(map[string]*provider.Record, 0),
	}
}

// Create implements Provider interface.
func (m *Mock) Create(r *provider.Record) error {
	if _, exists := m.recordsMap[r.Name]; exists {
		return provider.ErrExists
	}

	rCopy := *r
	m.recordsList = append(m.recordsList, &rCopy)
	m.recordsMap[r.Name] = &rCopy
	return nil
}

// Update implements Provider interface.
func (m *Mock) Update(r *provider.Record) error {
	record, found := m.recordsMap[r.Name]
	if !found {
		return provider.ErrNotFound
	}

	*record = *r
	return nil
}

// Get implements Provider interface.
func (m *Mock) Get(domain string) ([]*provider.Record, error) {
	return m.recordsList, nil
}
