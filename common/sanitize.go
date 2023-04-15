package common

import "strings"

func SanitizeStringForSql(str string) string {
	// TODO -- this is not acceptable for production
	s := strings.ReplaceAll(str, "'", "''")
	return s
}
