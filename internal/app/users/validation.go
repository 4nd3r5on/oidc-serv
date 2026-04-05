package users

import (
	"fmt"
	"unicode"

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
		return fmt.Errorf("username must be between %d and %d characters", usernameMinLen, usernameMaxLen)
	}
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf("username may only contain letters, digits and underscores")
		}
	}
	return nil
}

func validateLocale(locale string) error {
	if _, err := language.Parse(locale); err != nil {
		return fmt.Errorf("invalid locale %q: %w", locale, err)
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < passwordMinLen {
		return fmt.Errorf("password must be at least %d characters", passwordMinLen)
	}
	return nil
}
