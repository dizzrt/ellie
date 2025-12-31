package ptrconv

func Ptr[T any](v T) *T {
	return &v
}

func Val[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}

func ValWithDefault[T any](v *T, defaultValue T) T {
	if v == nil {
		return defaultValue
	}

	return *v
}
