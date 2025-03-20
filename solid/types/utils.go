package types

import "strings"

func normalizeName(name string) string {
	trimmed := strings.Join(strings.Fields(name), "-")
	return strings.ToLower(trimmed)
}

func NormalizeID(s string) string {
	return normalizeName(s)
}
