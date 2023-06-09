package common

import "os"

func Dedupe[T comparable](arr []T) []T {
	keys := make(map[T]bool)
	list := []T{}

	for _, entry := range arr {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func BToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

/************************************************************************************/
// Get a value from environment variables
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
