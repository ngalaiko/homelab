package mock

import (
	"net"

	"github.com/ngalayko/dyn-dns/app/fetcher"
)

// Mock is a mock fetcher.
type Mock struct {
	IP net.IP
}

// Fetch implements Fetcher interface.
func (m *Mock) Fetch() (net.IP, error) {
	if m.IP.String() == "<nil>" {
		return m.IP, fetcher.ErrUnavailable
	}
	return m.IP, nil
}
