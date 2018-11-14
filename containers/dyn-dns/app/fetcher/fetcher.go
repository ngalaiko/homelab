package fetcher

import (
	"net"
)

// Fetcher is used to fetch public ip.
type Fetcher interface {
	Fetch() (net.IP, error)
}
