// Package errs provides domain-neutral error sentinels and helpers.
package errs

import "errors"

var (
	ErrNotImplemented   = errors.New("not implemented")
	ErrRemoteServiceErr = errors.New("remote service error")
	ErrRateLimited      = errors.New("rate limited")

	ErrInvalidArgument = errors.New("invalid argument")
	ErrMissingArgument = errors.New("missing argument")
	ErrOutOfRange      = errors.New("out of range")

	ErrPermissionDenied = errors.New("permission denied")
	ErrUnauthorized     = errors.New("unauthorized")

	ErrExists   = errors.New("already exists")
	ErrNotFound = errors.New("not found")
	ErrOutdated = errors.New("outdated")
)

// IsAny reports whether err matches any of the given reference errors via errors.Is.
func IsAny(err error, references ...error) bool {
	for _, reference := range references {
		if errors.Is(err, reference) {
			return true
		}
	}
	return false
}

// Rewrap creates a new error with the given message and transfers the sentinels
// and response data from from onto it. Use this at security boundaries to
// replace an internal error message with a user-safe one without losing the
// HTTP-mapping metadata (status-code sentinels, extra response fields).
//
//	// Repository returns: errs.Mark(pgx.ErrNoRows, errs.ErrNotFound)
//	// App layer replaces the internal message but keeps the 404 sentinel:
//	return errs.Rewrap("user not found", err)
func Rewrap(msg string, from error) error {
	result := errors.New(msg)

	if m, ok := errors.AsType[*markedError](from); ok {
		result = Mark(append([]error{result}, m.sentinels...)...)
	}
	if rd, ok := errors.AsType[*respDataError](from); ok {
		result = NewRespData(result, rd.respData)
	}

	return result
}
