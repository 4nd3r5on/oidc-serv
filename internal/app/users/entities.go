package users

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID
	Locale       string
	Username     string
	PasswordHash []byte
}
