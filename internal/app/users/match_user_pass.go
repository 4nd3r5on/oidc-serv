package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials is returned when username or password do not match.
// A single sentinel is used for both cases to prevent username enumeration.
// Pre-marked with [errs.ErrUnauthorized] so the API layer maps it to 401.
var ErrInvalidCredentials = errs.Mark(errors.New("invalid credentials"), errs.ErrUnauthorized)

type MatchUserPassRepo interface {
	GetterByUsername
}

type MatchUserPass struct {
	Users  MatchUserPassRepo
	Logger *slog.Logger
}

func NewMatchUserPass(users MatchUserPassRepo, logger *slog.Logger) *MatchUserPass {
	if logger == nil {
		logger = slog.Default()
	}
	return &MatchUserPass{Users: users, Logger: logger}
}

// MatchUserPass looks up the user by username and verifies the password.
// Returns ErrInvalidCredentials for both unknown usernames and wrong passwords
// to prevent username enumeration.
func (m *MatchUserPass) MatchUserPass(ctx context.Context, username, password string) (*User, error) {
	user, err := m.Users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		m.Logger.ErrorContext(ctx, "match user pass: repository error", "error", err)
		return nil, fmt.Errorf("repository: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrInvalidCredentials
		}
		m.Logger.ErrorContext(ctx, "bcrypt compare failed", "error", err)
		return nil, fmt.Errorf("compare password: %w", err)
	}

	return user, nil
}
