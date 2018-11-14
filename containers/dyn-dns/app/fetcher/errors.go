package fetcher

import "fmt"

// List of common errors.
var (
	ErrUnavailable = fmt.Errorf("public ip unavailable.")
)
