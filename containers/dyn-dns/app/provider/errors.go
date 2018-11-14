package provider

import "fmt"

// List of common errors.
var (
	ErrNotFound = fmt.Errorf("not found")
	ErrExists   = fmt.Errorf("exists")
)
