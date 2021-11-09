package genshinartis

import (
	"fmt"
	"math/rand"
)

const MaxSubstats = 4

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

type Artifact struct {
	Set      artifactSet
	Slot     artifactSlot
	MainStat artifactStat
	SubStats [MaxSubstats]*ArtifactSubstat
}

func (a *Artifact) randomizeSet(options ...artifactSet) {
	a.Set = options[rand.Intn(len(options))]
}

func (a *Artifact) randomizeSlot() {
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
	if rand.Float32() <= 0.25 {
		numRolls++ // starts with 4 subs (25% chance)
	}

	a.SubStats = [MaxSubstats]*ArtifactSubstat{}
	possibleStats := weightedSubstats(a.MainStat)
	var subs [MaxSubstats]artifactStat
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
	artifact.randomizeSet(allArtifactSets...)
	artifact.randomizeSlot()
	artifact.ranzomizeMainStat()
	artifact.randomizeSubstats()
	return &artifact
}

func RandomArtifactOfSlot(slot artifactSlot) *Artifact {
	var artifact Artifact
	artifact.randomizeSet(allArtifactSets...)
	artifact.Slot = slot
	artifact.ranzomizeMainStat()
	artifact.randomizeSubstats()
	return &artifact
}

func RandomArtifactFromDomain(setA, setB string) *Artifact {
	var artifact Artifact
	artifact.randomizeSet(artifactSet(setA), artifactSet(setB))
	artifact.randomizeSlot()
	artifact.ranzomizeMainStat()
	artifact.randomizeSubstats()
	return &artifact
}
