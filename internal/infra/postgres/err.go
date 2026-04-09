package postgres

import (
	"errors"
	"strings"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mapErr translates postgres-specific errors to domain sentinel errors.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.Mark(err, errs.ErrNotFound)
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505": // unique_violation
			err = errs.Mark(err, errs.ErrExists)
			field, ok := fieldFromConstraint(pgErr.ConstraintName)
			if ok {
				err = errs.NewRespData(err, map[string]any{"conflict_field": field})
			}
			return err
		}
	}
	return err
}

// fieldFromConstraint extracts a human-readable field name from a postgres
// constraint name following the <table>_<field>_<suffix> naming convention.
// e.g. "users_username_idx" → "username"
func fieldFromConstraint(constraint string) (string, bool) {
	first := strings.Index(constraint, "_")
	last := strings.LastIndex(constraint, "_")
	if first == last {
		return "", false
	}
	return constraint[first+1 : last], true
}
