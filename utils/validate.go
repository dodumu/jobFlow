package utils

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MinPasswordLenght = 8

	MaxPasswordBytes = 72
)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return errors.New("please enter a valid email address")
	}

	at := strings.LastIndex(email, "@")
	if !strings.Contains(email[at+1:], ".") {
		return errors.New("Please enter a valid email address")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) > MaxPasswordBytes {
		return errors.New("Password is too long (maximum 72 characters).")
	}

	if utf8.RuneCountInString(password) < MinPasswordLenght {
		return errors.New("Password must be at least 8 characters long.")
	}

	var hasLetter, hasDigit bool

	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return errors.New("Password must contain at least on letter and one number")
	}
	return nil
}
