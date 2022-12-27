package util

// Transform converts a slice to another slice using the provided function.
func Transform[T any, V any](initial []T, transform func(T) V) []V {
	if len(initial) == 0 {
		return nil
	}
	result := make([]V, 0, len(initial))
	for _, item := range initial {
		result = append(result, transform(item))
	}
	return result
}
