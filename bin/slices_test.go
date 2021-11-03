package main

import (
	"math/rand"
	"testing"
)

func Test_extractRandomStringFromSlice(t *testing.T) {
	rand.Seed(666)
	slice := []string{"A", "B", "C", "D", "E"}

	if got := extractRandomStringFromSlice(&slice); got != "C" {
		t.Errorf("extractRandomStringFromSlice() = %v, want %v", got, "C")
	}
	if len(slice) != 4 {
		t.Errorf("expected len 4, got: %v", len(slice))
	}

	if got := extractRandomStringFromSlice(&slice); got != "B" {
		t.Errorf("extractRandomStringFromSlice() = %v, want %v", got, "B")
	}
	if len(slice) != 3 {
		t.Errorf("expected len 3, got: %v", len(slice))
	}
}
