package rngx

import (
	"math/rand"
	"testing"
)

func Test_PickAndRemove(t *testing.T) {
	rand.Seed(666)
	slice := []string{"A", "B", "C", "D", "E"}

	if got := PickAndRemove(&slice); got != "C" {
		t.Errorf("PickAndRemove() = %v, want %v", got, "C")
	}
	if len(slice) != 4 {
		t.Errorf("expected len 4, got: %v", len(slice))
	}

	if got := PickAndRemove(&slice); got != "B" {
		t.Errorf("PickAndRemove() = %v, want %v", got, "B")
	}
	if len(slice) != 3 {
		t.Errorf("expected len 3, got: %v", len(slice))
	}
}

func Test_PickAndRemove_Empty(t *testing.T) {
	slice := []string{}
	if got := PickAndRemove(&slice); got != "" {
		t.Errorf("PickAndRemove() = %v, want %v", got, "")
	}
}
