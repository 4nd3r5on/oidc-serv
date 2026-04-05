package users

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type MatchUserPassRepo interface {
	GetterByUsername
}

type MatchUserPass struct {
	Users MatchUserPassRepo
}

// Match looks up the user by username and verifies the password.
// Returns ErrInvalidCredentials for both unknown usernames and wrong passwords
// to prevent username enumeration.
func (m *MatchUserPass) Match(ctx context.Context, username, password string) (*User, error) {
	user, err := m.Users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("repository: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("compare password: %w", err)
	}

	return user, nil
}
