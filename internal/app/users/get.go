package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
	"github.com/google/uuid"
)

// GetRes is the safe, hash-free view of a user returned to callers.
type GetRes struct {
	ID       uuid.UUID
	Locale   string
	Username string
}

type GetByID struct {
	Users  GetterByID
	Logger *slog.Logger
}

func NewGetByID(users GetterByID, logger *slog.Logger) *GetByID {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetByID{Users: users, Logger: logger}
}

func (g *GetByID) Get(ctx context.Context, id uuid.UUID) (*GetRes, error) {
	user, err := g.Users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.Rewrap("user not found", err)
		}
		g.Logger.ErrorContext(ctx, "get user by id: repository error", "error", err)
		return nil, fmt.Errorf("repository: %w", err)
	}
	return userToGetRes(user), nil
}

type GetByUsername struct {
	Users  GetterByUsername
	Logger *slog.Logger
}

func NewGetByUsername(users GetterByUsername, logger *slog.Logger) *GetByUsername {
	if logger == nil {
		logger = slog.Default()
	}
	return &GetByUsername{Users: users, Logger: logger}
}

func (g *GetByUsername) Get(ctx context.Context, username string) (*GetRes, error) {
	user, err := g.Users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.Rewrap("user not found", err)
		}
		g.Logger.ErrorContext(ctx, "get user by username: repository error", "error", err)
		return nil, fmt.Errorf("repository: %w", err)
	}
	return userToGetRes(user), nil
}

func userToGetRes(u *User) *GetRes {
	return &GetRes{
		ID:       u.ID,
		Username: u.Username,
		Locale:   u.Locale,
	}
}
