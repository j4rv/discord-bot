package genshinartis

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestRandomArtifactFromDomain(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	var set1Count, set2Count int
	set1, set2 := "A", "B"

	// Generate 1000 artifacts from two sets
	for i := 0; i < 1000; i++ {
		art := RandomArtifactFromDomain(set1, set2)
		if art.Set == artifactSet(set1) {
			set1Count++
		} else if art.Set == artifactSet(set2) {
			set2Count++
		} else {
			t.Error("Unexpected artifact set: " + art.Set)
		}
	}

	// Then check that the chances of getting an artifact from either set is ~50%
	if set1Count < 450 || set1Count > 550 {
		t.Error("Too many or too few artifacts from set 1: " + strconv.Itoa(set1Count))
	}
	if set2Count < 450 || set2Count > 550 {
		t.Error("Too many or too few artifacts from set 1: " + strconv.Itoa(set2Count))
	}
}

// Just for theorycrafting
func TestRandomArtifact(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	var count int
	for i := 0; i < 1000000; i++ {
		art := RandomArtifact()
		var cv float32
		for _, sub := range art.SubStats {
			if sub.Stat == CritRate {
				cv += sub.Value * 2
			}
			if sub.Stat == CritDmg {
				cv += sub.Value
			}
		}
		if cv >= 50 {
			t.Log("------------")
			t.Log(art)
			t.Log(cv)
			count++
		}
	}
	t.Log("total:", count)
}
