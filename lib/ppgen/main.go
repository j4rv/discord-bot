package ppgen

import (
	"math/rand"
	"time"
)

const minLength = 1
const maxLength = 14

// FIXME:
// Since map's order is arbitrary, this generator will generate different penises even when given the same seed
// Change to a custom-made ordered map or something idk

var leftPPHeads = map[string]int{
	"C": 100,
	"O": 60,
	"c": 40,
	"(": 20,
	"<": 10,
}

var leftPPBalls = map[string]int{
	"8": 100,
	"3": 80,
	"B": 60,
	"}": 10,
	"]": 10,
}

var rightPPHeads = map[string]int{
	"D": 100,
	"϶": 80,
	"O": 60,
	")": 20,
	">": 10,
}

var rightPPBalls = map[string]int{
	"8": 100,
	"E": 20,
	"}": 10,
	"]": 10,
	"∑": 5,
}

var ppBodies = map[string]int{
	"=":   100,
	"≈":   40,
	"-":   10,
	"\\~": 10,
}

func NewPenis() string {
	return NewPenisWithSeed(time.Now().Unix())
}

func NewPenisWithSeed(seed int64) string {
	rng := rand.New(rand.NewSource(seed))
	facingLeft := rng.Float64() <= 0.5

	if facingLeft {
		return newPenisFacingLeft(rng)
	} else {
		return newPenisFacingRight(rng)
	}
}

func newPenisFacingLeft(rng *rand.Rand) string {
	length := rng.Intn(maxLength-minLength) + minLength
	head := weightedRand(rng, leftPPHeads)
	body := weightedRand(rng, ppBodies)
	balls := weightedRand(rng, leftPPBalls)

	penis := head
	for i := 0; i < length; i++ {
		penis += body
	}
	penis += balls

	return penis
}

func newPenisFacingRight(rng *rand.Rand) string {
	length := rng.Intn(maxLength-minLength) + minLength
	balls := weightedRand(rng, rightPPBalls)
	body := weightedRand(rng, ppBodies)
	head := weightedRand(rng, rightPPHeads)

	penis := balls
	for i := 0; i < length; i++ {
		penis += body
	}
	penis += head

	return penis
}
