package utils

import "fmt"

func Ptr[T any](v T) *T {
	return &v
}

func FormatDuration(ms int) string {
	seconds := ms / 1000
	minutes := seconds / 60
	seconds %= 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
