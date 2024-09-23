package nimble

// Unique returns a slice from the original containing only unique elements
func Unique[T comparable](values ...T) []T {
	filtered := []T{}
	s := NewSet[T]()
	for _, v := range values {
		if !s.Has(v) {
			filtered = append(filtered, v)
			s.Add(v)
		}
	}

	return filtered
}

// Filter returns the filtered items from the input set, using the filter function
func Filter[T any](f func(index int, v T) bool, values ...T) []T {
	filtered := []T{}
	for index, v := range values {
		if f(index, v) {
			filtered = append(filtered, v)
		}
	}

	return filtered
}

// Map returns the original slice passed through the mapping function
func Map[T any, U any](f func(index int, v T) U, values ...T) []U {
	output := make([]U, len(values))
	for idx, v := range values {
		output[idx] = f(idx, v)
	}

	return output
}

// Or returns the orValue when value is the zero value for a given type
func Or[T comparable](value T, orValue T) T {
	var zero T
	if zero == value {
		return orValue
	}

	return value
}
