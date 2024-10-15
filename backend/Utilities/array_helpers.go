package util

// dst must be already allocated
func Copy2DArray[T any](dst [][]T, src [][]T) {
	for i := range src {
		dst[i] = make([]T, len(src[i]))
		copy(dst[i], src[i])
	}

}
