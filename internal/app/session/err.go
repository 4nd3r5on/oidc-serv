package session

import (
	"errors"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
)

// ErrSessionExpired is returned when a session token is valid but has passed
// its expiry time.
var ErrSessionExpired = errs.Mark(errors.New("session expired"), errs.ErrUnauthorized)

// ErrInvalidSession is returned when a session token is not found.
// Using a single error for "not found" prevents leaking which tokens exist.
var ErrInvalidSession = errs.Mark(errors.New("invalid or expired session"), errs.ErrUnauthorized)

var ErrInvalidSessionKey = errs.Mark(errors.New("invalid session key"), errs.ErrInvalidArgument)
