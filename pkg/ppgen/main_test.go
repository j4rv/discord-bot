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

// to check at what hour it changes
func TestNewPenisWithSeed2(t *testing.T) {
	for i := int64(0); i < 48*2; i++ {
		seed := (time.Now().Unix() + 15*60*i) / (60 * 60 * 24)
		penis := NewPenisWithSeed(seed)
		t.Logf("Seed: %d, Penis: %s, i: %d", seed, penis, i)
	}
}
