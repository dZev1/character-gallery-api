package auth

type AuthStore interface {
	ValidateAPIKey(keyHash string) (bool, error)
	UpdateLastUsed(keyHash string) error
	CreateAPIKey(name string) (string, error)
}