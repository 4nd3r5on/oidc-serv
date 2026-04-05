package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Delete struct {
	Users Deleter
}

func (d *Delete) Delete(ctx context.Context, id uuid.UUID) error {
	if err := d.Users.Delete(ctx, id); err != nil {
		return fmt.Errorf("repository: %w", err)
	}
	return nil
}
