package auth

import "time"

type User struct {
	ID           uint64
	Username     string
	PasswordHash string
	CreatedAt    *time.Time
}
