// Package engine contains interfaces for different git providers.
package engine

func transform[T any, V any](initial []T, transform func(T) V) []V {
	if len(initial) == 0 {
		return nil
	}
	result := make([]V, 0, len(initial))
	for _, item := range initial {
		result = append(result, transform(item))
	}
	return result
}
