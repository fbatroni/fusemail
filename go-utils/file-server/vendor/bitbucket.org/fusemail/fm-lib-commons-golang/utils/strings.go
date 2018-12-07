package utils

// Provides string utilities.

import (
	"fmt"
	"regexp"
	"strings"
)

// IndexOfString returns the index of str in the list,
// otherwise -1 if not found.
func IndexOfString(list []string, str string) int {
	for i, each := range list {
		if each == str {
			return i
		}
	}
	return -1
}

// ContainsString returns whether str is included in the list or not.
func ContainsString(list []string, str string) bool {
	i := IndexOfString(list, str)
	return i > -1
}

// StrV returns the string "%v" representation for any interface.
func StrV(any interface{}) string {
	return fmt.Sprintf("%v", any)
}

// StrPlus returns an expanded string representation for any interface.
// It avoids Stringer's infinite recursion issue for some cases.
func StrPlus(any interface{}) string {
	return fmt.Sprintf("%+v", any)
}

// SanitizeSpaces returns a new string with sanitized spaces (trim and CRLF),
// susceptible for comparison against another one using this same func.
func SanitizeSpaces(str string) string {
	re := regexp.MustCompile(`\r?\n`)
	return re.ReplaceAllString(strings.TrimSpace(str), " ")
}
