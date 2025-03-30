package types

import (
	"strconv"
	"strings"
)

func normalizeName(name string) string {
	trimmed := strings.Join(strings.Fields(name), "-")

	return strings.ToLower(trimmed)
}

func NormalizeID(s string) string {
	return normalizeName(s)
}

func ParseVal(s string) any {
	trimmed := strings.TrimSpace(s)

	f, err := strconv.ParseFloat(trimmed, 64)
	if err == nil {
		if f == float64(int64(f)) {
			return int64(f)
		}

		return f
	}

	return s
}
