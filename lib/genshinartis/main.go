package genshinartis

import (
	"fmt"
	"math/rand"
)

const MaxSubstats = 4

type Artifact struct {
	Slot     artifactSlot
	MainStat artifactStat
	SubStats [MaxSubstats]*ArtifactSubstat
}

type ArtifactSubstat struct {
	Stat  artifactStat
	Rolls int
	Value float32
}

func (s *ArtifactSubstat) randomizeValue() {
	s.Value = 0
	for i := 0; i < s.Rolls; i++ {
		s.Value = s.Value + s.Stat.RandomRollValue()
	}
}

func (s *ArtifactSubstat) String() string {
	return fmt.Sprintf("%s: %.1f", s.Stat, s.Value)
}

func (a *Artifact) randomizeType() {
	a.Slot = artifactSlot(rand.Intn(5))
}

func (a *Artifact) ranzomizeMainStat() {
	switch a.Slot {
	case SlotFlower:
		a.MainStat = HP
	case SlotPlume:
		a.MainStat = ATK
	case SlotSands:
		a.MainStat = weightedRand(sandsWeightedStats)
	case SlotGoblet:
		a.MainStat = weightedRand(gobletWeightedStats)
	case SlotCirclet:
		a.MainStat = weightedRand(circletWeightedStats)
	}
}

func (a *Artifact) randomizeSubstats() {
	numRolls := 3 + 5 // starts with 3 subs by default
	if rand.Float32() <= 0.5 {
		numRolls++ // starts with 4 subs (50% chance)
	}

	a.SubStats = [MaxSubstats]*ArtifactSubstat{}
	possibleStats := weightedSubstats(a.MainStat)
	var subs [4]artifactStat
	for i := 0; i < numRolls; i++ {
		// First 4 rolls
		if i < MaxSubstats {
			stat := weightedRand(possibleStats)
			subs[i] = stat
			a.SubStats[i] = &ArtifactSubstat{Stat: stat, Rolls: 1}
			delete(possibleStats, stat)
		} else {
			// Rest of rolls
			index := rand.Intn(MaxSubstats)
			a.SubStats[index].Rolls += 1
		}
	}

	for _, substat := range a.SubStats {
		substat.randomizeValue()
	}
}

func RandomArtifact() *Artifact {
	var artifact Artifact
	artifact.randomizeType()
	artifact.ranzomizeMainStat()
	artifact.randomizeSubstats()
	return &artifact
}

func RandomArtifactOfSlot(slot artifactSlot) *Artifact {
	var artifact Artifact
	artifact.Slot = slot
	artifact.ranzomizeMainStat()
	artifact.randomizeSubstats()
	return &artifact
}