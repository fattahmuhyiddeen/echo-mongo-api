package logic

import (
	"regexp"
	"strings"

	config "../config"
)

// IsValidEmail email format
func IsValidEmail(email string) bool {
	if config.IsProduction() && strings.Contains(email, "yopmail") {
		return false
	}

	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	return re.MatchString(email)
}
