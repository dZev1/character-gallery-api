// Package uuidv7 wraps google/uuid's time-ordered UUID v7 generation so the
// rest of the codebase depends on a single entry point. V7 embeds a Unix
// millisecond timestamp in the high 48 bits, giving monotonic ordering that
// keeps b-tree index locality high (unlike random v4).
package uuidv7

import "github.com/google/uuid"

// New generates a new time-ordered UUID (version 7).
func New() (uuid.UUID, error) {
	return uuid.NewV7()
}

// Must is like New but panics on error. Use only for variables whose
// generation cannot fail (e.g. keys used to satisfy a non-nullable PK).
func Must() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
