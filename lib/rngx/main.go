package rngx

import "math/rand"

func Pick[A any](slice []A) A {
	var picked A
	if len(slice) == 0 {
		return picked
	}
	return slice[rand.Intn(len(slice))]
}

func PickWithSource[A any](slice []A, rng *rand.Rand) A {
	var picked A
	if len(slice) == 0 {
		return picked
	}
	return slice[rng.Intn(len(slice))]
}
