package utils

import (
	"errors"
	"strings"
	"unicode"
)

func ValidateName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return errors.New("name cannot be empty")
	}

	if len(name) < 2 {
		return errors.New("name must be atleast two characters long")
	}

	if len(name) > 100 {
		return errors.New("name must not exceed 100 characters")
	}

	for _, char := range name {
		if unicode.IsLetter(char) {
			continue
		}

		if unicode.IsDigit(char) {
			continue
		}

		if char == '.' {
			continue
		}

		if char == ' ' {
			continue
		}

		if char == '-' {
			continue
		}

		if char == '\'' {
			continue
		}

		return errors.New("name can only contain letters, numbers, periods, spaces, hyphens and apostrophes")
	}
	return nil
}
