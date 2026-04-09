package errs

import "errors"

// markedError wraps a base error with one or more sentinels for matching.
type markedError struct {
	err       error
	sentinels []error
}

// Error returns the original error message, satisfying the error interface.
func (m *markedError) Error() string {
	return m.err.Error()
}

// Unwrap returns the underlying error so errors.As and other wrapping logic works.
func (m *markedError) Unwrap() error {
	return m.err
}

// Is checks if the target matches any of the assigned sentinels.
func (m *markedError) Is(target error) bool {
	for _, s := range m.sentinels {
		if errors.Is(s, target) || s == target {
			return true
		}
	}
	return false
}

// Mark attaches one or more sentinel errors to err so that errors.Is matches
// them without changing err's message or breaking the Unwrap chain.
//
// The first argument is the subject error; all subsequent arguments are
// sentinels. Returns nil if err is nil. Returns err unchanged if no sentinels
// are provided.
//
//	err = errs.Mark(err, errs.ErrNotFound)
//	errors.Is(err, errs.ErrNotFound) // true
//	err.Error()                       // original message unchanged
func Mark(errs ...error) error {
	if len(errs) == 0 || errs[0] == nil {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return &markedError{
		err:       errs[0],
		sentinels: errs[1:],
	}
}
