package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/4nd3r5on/oidc-serv/internal/app/users"
	"github.com/4nd3r5on/oidc-serv/pkg/db"
)

// UserRepo is a PostgreSQL-backed implementation of the user repository
// interfaces defined in the users package ([users.Creator], [users.GetterByID],
// [users.GetterByUsername], [users.Updater], [users.Deleter]).
type UserRepo struct {
	q *db.Queries
}

// NewUserRepo returns a UserRepo backed by the given sqlc query set.
func NewUserRepo(q *db.Queries) *UserRepo {
	return &UserRepo{q: q}
}

// Create inserts a new user row and returns the generated UUID.
// Returns an error marked with [errs.ErrExists] if the username is already in use.
func (r *UserRepo) Create(ctx context.Context, opts users.CreateDBOpts) (uuid.UUID, error) {
	id, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Username:     opts.Username,
		PasswordHash: opts.PasswordHash,
		Locale:       opts.Locale,
	})
	if err != nil {
		return uuid.Nil, mapErr(err)
	}
	return uuid.UUID(id.Bytes), nil
}

// GetByID fetches a user by primary key.
// Returns an error marked with [errs.ErrNotFound] when no row matches.
func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	row, err := r.q.GetUserByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return nil, mapErr(err)
	}
	return rowToUser(row.ID, row.Username, row.PasswordHash, row.Locale), nil
}

// GetByUsername fetches a user by their unique username.
// Returns an error marked with [errs.ErrNotFound] when no row matches.
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*users.User, error) {
	row, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, mapErr(err)
	}
	return rowToUser(row.ID, row.Username, row.PasswordHash, row.Locale), nil
}

// Update applies a partial update to the user identified by id.
// Only fields set in opts are written; nil pointer fields and a nil Password
// slice are left unchanged via COALESCE in the underlying query.
// Returns an error marked with [errs.ErrExists] if the new username conflicts.
func (r *UserRepo) Update(ctx context.Context, id uuid.UUID, opts users.UpdateDBOpts) error {
	params := db.UpdateUserParams{
		ID:           pgtype.UUID{Bytes: id, Valid: true},
		PasswordHash: opts.Password,
	}
	if opts.Username != nil {
		params.Username = pgtype.Text{String: *opts.Username, Valid: true}
	}
	if opts.Locale != nil {
		params.Locale = pgtype.Text{String: *opts.Locale, Valid: true}
	}
	return mapErr(r.q.UpdateUser(ctx, params))
}

// Delete removes the user row with the given id.
func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return mapErr(r.q.DeleteUser(ctx, pgtype.UUID{Bytes: id, Valid: true}))
}

// Exists reports whether a user with the given id exists.
func (r *UserRepo) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.q.UserExists(ctx, pgtype.UUID{Bytes: id, Valid: true})
}
