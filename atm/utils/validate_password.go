package utils

import "regexp"

func ValidatePassword(password string) bool {

	if len(password) < 8 {
		return false
	}

	upperRegex := regexp.MustCompile(`[A-Z]`)
	if !upperRegex.MatchString(password) {
		return false
	}

	specialRegex := regexp.MustCompile(`[^a-zA-Z0-9]`)
	if !specialRegex.MatchString(password) {
		return false
	}

	return true
}
