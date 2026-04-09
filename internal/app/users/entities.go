package users

import "github.com/google/uuid"

// User is the internal user representation.
// PasswordHash is never included in API responses; use [GetRes] at the boundary.
type User struct {
	ID           uuid.UUID
	Locale       string
	Username     string
	PasswordHash []byte
}
