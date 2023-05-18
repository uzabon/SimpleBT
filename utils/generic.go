package utils

func Of[T any](d T) *T {
	return &d
}
