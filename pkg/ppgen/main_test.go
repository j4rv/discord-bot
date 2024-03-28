package ppgen

import (
	"testing"
	"time"
)

func TestNewPenisWithSeed(t *testing.T) {
	for i := int64(0); i < 300; i++ {
		seed := time.Now().UnixNano() + i
		penis := NewPenisWithSeed(seed)
		t.Logf("Seed: %d, Penis: %s", seed, penis)
	}
}
