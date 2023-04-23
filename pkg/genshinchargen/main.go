package genshinchargen

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
)

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

var weapon = NewWeightedSlice(map[string]int{
	"Sword":            100,
	"Claymore":         100,
	"Polearm":          100,
	"Catalyst":         100,
	"Bow":              100,
	"Gun":              1,
	"Scythe":           1,
	"Brawler":          1,
	"Whip":             1,
	"Sword and Shield": 1,
})

var rarity = NewWeightedSlice(map[string]int{
	"6*": 1,
	"5*": 100,
	"4*": 100,
	"3*": 1,
})

var models = NewWeightedSlice(map[string]int{
	"Tall male":     50,
	"Tall female":   50,
	"Medium male":   100,
	"Medium female": 100,
	"Short male":    40,
	"Short female":  40,
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
	"Off-field DPS":       7,
	"Buffer":              5,
	"Healer":              5,
	"Shielder":            5,
	"Phys on-field DPS":   3,
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
	"can group enemies":                         5,
	"has damage resistance buffs":               2,
	"can shred defense":                         2,
})

var weaknesses = NewWeightedSlice(map[string]int{
	"has energy issues":                           8,
	"very hard to play":                           8,
	"needs constellations to be good":             8,
	"selfish and needs a lot of field time":       8,
	"extremely fragile":                           8,
	"has scuffed ICDs":                            8,
	"has very long cooldowns":                     8,
	"has shitty multipliers":                      8,
	"the kit is circle impact":                    8,
	"it's purely single target":                   8,
	"consumes a lot of stamina to play optimally": 5,
	"needs resistance to interruption to be good": 5,
	"doesn't create particles":                    5,
	"can't crit":                                  2,
})

type GeneratedCharacter struct {
	name     string
	rarity   string
	element  string
	weapon   string
	model    string
	scaling  string
	role     string
	strength string
	weakness string
}

func (c GeneratedCharacter) PrettyString() string {
	return fmt.Sprintf("%s is a %s %s character.\nWeapon: %s.\nModel: %s.\nKit: %s, scales with %s, %s but %s.",
		c.name, c.rarity, c.element, c.weapon, c.model, c.role, c.scaling, c.strength, c.weakness)
}

func NewChar(name string, seedSalt int64) GeneratedCharacter {
	var result GeneratedCharacter
	result.name = name

	rng := rand.New(rand.NewSource(generateSeedFromName(name) + seedSalt))
	result.element = elements.Random(rng)
	result.rarity = rarity.Random(rng)
	result.weapon = weapon.Random(rng)
	result.model = models.Random(rng)
	result.scaling = scaling.Random(rng)
	result.role = roles.Random(rng)
	result.strength = strengths.Random(rng)
	result.weakness = weaknesses.Random(rng)

	return result
}

func generateSeedFromName(name string) int64 {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	hash := sha256.New()
	hash.Write([]byte(strings.ToLower(name)))
	return int64(binary.LittleEndian.Uint64(hash.Sum(nil)))
}
