package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           UserID     `db:"id" json:"id"`
	Username     string     `db:"username" json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Role         string     `db:"role" json:"role"`
	CreatedAt    *time.Time `db:"created_at" json:"-"`
}

type UserID uuid.UUID

func (id UserID) String() string {
	return uuid.UUID(id).String()
}

func (id UserID) MarshalText() ([]byte, error) {
	return uuid.UUID(id).MarshalText()
}

func (id *UserID) UnmarshalText(data []byte) error {
	var parsed uuid.UUID
	if err := parsed.UnmarshalText(data); err != nil {
		return err
	}
	*id = UserID(parsed)
	return nil
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type UserResponse struct {
	ID        UserID    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
