package validator

import (
	"regexp"
	"unicode"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 255 {
		return false
	}
	return emailRegex.MatchString(email)
}

func IsValidPassword(password string) bool {
	if len(password) < 6 || len(password) > 72 {
		return false
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Пароль должен содержать все: заглавные, строчные, цифры и спецсимволы
	return hasUpper && hasLower && hasNumber && hasSpecial
}

func IsValidName(name string) bool {
	return len(name) >= 2 && len(name) <= 100
}

func IsValidBio(bio string) bool {
	return len(bio) <= 200
}
