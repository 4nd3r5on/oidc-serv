package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type GetRes struct {
	ID       uuid.UUID
	Locale   string
	Username string
}

type GetByID struct {
	Users GetterByID
}

func (g *GetByID) Get(ctx context.Context, id uuid.UUID) (*GetRes, error) {
	user, err := g.Users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("repository: %w", err)
	}
	return userToGetRes(user), nil
}

type GetByUsername struct {
	Users GetterByUsername
}

func (g *GetByUsername) Get(ctx context.Context, username string) (*GetRes, error) {
	user, err := g.Users.GetByUsername(ctx, username)
	if err != nil {
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
