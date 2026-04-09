package users

import (
	"fmt"
	"unicode"

	"github.com/4nd3r5on/oidc-serv/pkg/errs"
	"golang.org/x/text/language"
)

const (
	usernameMinLen = 3
	usernameMaxLen = 32
	passwordMinLen = 8
)

func validateUsername(username string) error {
	l := len(username)
	if l < usernameMinLen || l > usernameMaxLen {
		return errs.Mark(
			fmt.Errorf("username must be between %d and %d characters", usernameMinLen, usernameMaxLen),
			errs.ErrInvalidArgument,
		)
	}
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return errs.Mark(
				fmt.Errorf("username may only contain letters, digits and underscores"),
				errs.ErrInvalidArgument,
			)
		}
	}
	return nil
}

func validateLocale(locale string) error {
	if _, err := language.Parse(locale); err != nil {
		return errs.Mark(fmt.Errorf("invalid locale %q", locale), errs.ErrInvalidArgument)
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < passwordMinLen {
		return errs.Mark(
			fmt.Errorf("password must be at least %d characters", passwordMinLen),
			errs.ErrInvalidArgument,
		)
	}
	return nil
}
