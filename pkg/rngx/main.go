package rngx

import "math/rand"

func Pick[A any](slice []A) A {
	var picked A
	if len(slice) == 0 {
		return picked
	}
	return slice[rand.Intn(len(slice))]
}

func PickAndRemove[A any](s *[]A) A {
	var picked A
	if len(*s) == 0 {
		return picked
	}
	i := int(rand.Float32() * float32(len(*s)))
	elem := (*s)[i]
	(*s)[i] = (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return elem
}

func PickWithSource[A any](slice []A, rng *rand.Rand) A {
	var picked A
	if len(slice) == 0 {
		return picked
	}
	return slice[rng.Intn(len(slice))]
}
