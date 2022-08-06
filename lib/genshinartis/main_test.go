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
	set1, set2 := "Emblem", "Shimenawa"

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

func TestRemoveTrashArtifacts(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	var artis []*Artifact
	set1, set2 := "Emblem", "Shimenawa"

	// Generate 1000 artifacts from two sets
	for i := 0; i < 10000; i++ {
		artis = append(artis, RandomArtifactFromDomain(set1, set2))
	}

	subs := map[artifactStat]float32{
		ATKP:           1,
		CritRate:       1,
		CritDmg:        1,
		EnergyRecharge: 0.5,
		ATK:            0.25,
	}
	filtered := RemoveTrashArtifacts(artis, subs, 5)
	for _, a := range filtered {
		t.Log(*a)
	}
}

// Just for theorycrafting
func TestTheorycrafting(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())

	set1, set2 := "DEF", "Heals"
	iterations := 5000
	artifactsAmount := 96
	wantedSet := set1
	wantedSlot := SlotCirclet
	wantedMainStat1, wantedMainStat2 := CritDmg, CritDmg
	subValue := map[artifactStat]float32{
		CritRate:       1,
		CritDmg:        1,
		DEFP:           0.9,
		ATKP:           0.8,
		EnergyRecharge: 0.4,
		DEF:            0.2,
		ATK:            0.1,
	}

	var avg float32
	for j := 0; j < iterations; j++ {
		var best *Artifact
		for i := 0; i < artifactsAmount; i++ {
			art := RandomArtifactFromDomain(set1, set2)
			if art.Set != artifactSet(wantedSet) {
				continue
			}
			if !(art.MainStat == wantedMainStat1 || art.MainStat == wantedMainStat2) || art.Slot != wantedSlot {
				continue
			}
			if best == nil || best.subsQuality(subValue) < art.subsQuality(subValue) {
				best = art
			}
		}
		if best == nil {
			continue
		}
		avg += float32(best.subsQuality(subValue))
	}
	t.Log(wantedSlot, wantedMainStat1, wantedMainStat2)
	t.Log("Average:", avg/float32(iterations))
}
