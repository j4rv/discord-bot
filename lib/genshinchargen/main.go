package genshinchargen

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
)

// wip

var elements = NewWeightedSlice(map[string]int{
	"Pyro":    100,
	"Hydro":   100,
	"Electro": 100,
	"Cryo":    100,
	"Anemo":   100,
	"Geo":     100,
	"Dendro":  100,
	"Abyss":   1,
	"Omni":    1,
})

var rarity = NewWeightedSlice(map[string]int{
	"5*": 10,
	"4*": 10,
})

var models = NewWeightedSlice(map[string]int{
	"tall male":     50,
	"tall female":   50,
	"medium male":   100,
	"medium female": 100,
	"short male":    40,
	"short female":  40,
})

var scaling = NewWeightedSlice(map[string]int{
	"ATK":             50,
	"HP":              20,
	"DEF":             10,
	"EM":              10,
	"Energy Recharge": 3,
	"Healing Bonus":   2,
	"Shield Strength": 1,
})

var roles = NewWeightedSlice(map[string]int{
	"On-field DPS":        10,
	"Off-field DPS":       5,
	"Buffer":              5,
	"Healer":              5,
	"Shielder":            5,
	"Healer and shielder": 2,
})

var strengths = NewWeightedSlice(map[string]int{
	"has good AOE":                              10,
	"has good elemental application":            10,
	"it's a great battery":                      10,
	"easy to build":                             10,
	"has very high damage":                      10,
	"has great vertical scaling":                10,
	"has amazing animations and visual effects": 10,
	"can snapshot buffs":                        8,
	"can shred resistances":                     5,
	"has damage resistance buffs":               2,
	"can shred defense":                         2,
})

var weaknesses = NewWeightedSlice(map[string]int{
	"extremely fragile":                           10,
	"has energy issues":                           10,
	"very hard to play":                           10,
	"needs constellations to be good":             10,
	"selfish and needs a lot of field time":       10,
	"has scuffed ICDs":                            8,
	"has very long cooldowns":                     8,
	"has shitty multipliers":                      8,
	"his kit is circlet impact":                   8,
	"needs resistance to interruption to be good": 5,
	"doesn't produce particles":                   2,
	"can't crit":                                  2,
})

type GeneratedCharacter struct {
	name     string
	rarity   string
	element  string
	model    string
	scaling  string
	role     string
	strength string
	weakness string
}

func (c GeneratedCharacter) String() string {
	return fmt.Sprintf("%s is a %s %s %s with a %s model and scales with %s, %s but %s.",
		c.name, c.rarity, c.element, c.role, c.model, c.scaling, c.strength, c.weakness)
}

func NewChar(name string) GeneratedCharacter {
	var result GeneratedCharacter
	result.name = name

	rng := rand.New(rand.NewSource(generateSeed(name)))
	result.element = elements.Random(rng)
	result.rarity = rarity.Random(rng)
	result.model = models.Random(rng)
	result.scaling = scaling.Random(rng)
	result.role = roles.Random(rng)
	result.strength = strengths.Random(rng)
	result.weakness = weaknesses.Random(rng)

	return result
}

func generateSeed(name string) int64 {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	hash := sha256.New()
	hash.Write([]byte(strings.ToLower(name)))
	return int64(binary.LittleEndian.Uint64(hash.Sum(nil)))
}
