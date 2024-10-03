package util

// dst must be already allocated
func Copy2DArrayBool(dst [][]bool, src [][]bool){ 
	for i := range src{
		dst[i] = make([]bool, len(src[i]))
		copy(dst[i], src[i])
	}

}