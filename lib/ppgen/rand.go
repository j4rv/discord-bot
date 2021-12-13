package ppgen

import (
	"log"
	"math/rand"
)

func weightedRand(rng *rand.Rand, runes map[string]int) string {
	sum := 0
	for _, weight := range runes {
		sum += weight
	}

	i := rng.Intn(sum)
	for value, weight := range runes {
		i -= weight
		if i < 0 {
			return value
		}
	}

	log.Println("fatal error in WeightedRand: should never reach this log")
	return ""
}
