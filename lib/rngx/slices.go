package rngx

import (
	"math/rand"
)

func ExtractRandomStringFromSlice(s *[]string) string {
	i := int(rand.Float32() * float32(len(*s)))
	elem := (*s)[i]
	(*s)[i] = (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return elem
}
