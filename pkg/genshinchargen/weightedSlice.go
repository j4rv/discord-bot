package genshinchargen

import (
	"log"
	"math/rand"
)

type WeightedSlice[V any] struct {
	Entries     []WeightedSliceEntry[V]
	TotalWeight int
}

func NewWeightedSlice[V comparable](input map[V]int) WeightedSlice[V] {
	res := WeightedSlice[V]{}
	for value, weight := range input {
		res.Entries = append(res.Entries, WeightedSliceEntry[V]{value, weight})
	}
	res.CalcWeight()
	return res
}

func (w *WeightedSlice[V]) CalcWeight() {
	sum := 0
	for _, e := range w.Entries {
		sum += e.Weight
	}
	w.TotalWeight = sum
}

func (w WeightedSlice[V]) Random(rng *rand.Rand) V {
	i := rng.Intn(w.TotalWeight)
	for _, e := range w.Entries {
		i -= e.Weight
		if i < 0 {
			return e.Value
		}
	}

	log.Println("fatal error in WeightedSlice.Random: should never reach this log")
	var zeroVal V
	return zeroVal
}

type WeightedSliceEntry[V any] struct {
	Value  V
	Weight int
}
