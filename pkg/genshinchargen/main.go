package genshinchargen

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
)

var elements = NewWeightedSlice(map[string]int{
	"Pyro":      100,
	"Hydro":     100,
	"Electro":   100,
	"Cryo":      100,
	"Anemo":     100,
	"Geo":       100,
	"Dendro":    100,
	"Abyss":     1,
	"Solar":     1,
	"Omni":      1,
	"Quantum":   1,
	"Imaginary": 1,
})

var weapon = NewWeightedSlice(map[string]int{
	"Sword":                   1000,
	"Claymore":                1000,
	"Polearm":                 1000,
	"Bow":                     800,
	"Bow (with a special CA)": 200,
	"Catalyst (ranged)":       700,
	"Catalyst (melee)":        300,
	"Dual swords":             5,
	"Gun":                     5,
	"Shotgun":                 2,
	"Scythe":                  2,
	"Brawler":                 2,
	"Whip":                    2,
	"Sword and Shield":        2,
})

var rarity = NewWeightedSlice(map[string]int{
	"7*": 1,
	"6*": 10,
	"5*": 1000,
	"4*": 1000,
	"3*": 10,
	"2*": 1,
})

var region = NewWeightedSlice(map[string]int{
	"Mondstadt":   100,
	"Liyue":       100,
	"Inazuma":     100,
	"Sumeru":      100,
	"Fontaine":    100,
	"Natlan":      100,
	"Snezhnaya":   100,
	"Khaenri'ah":  10,
	"Celestia":    10,
	"Enkanomiya":  5,
	"The Chasm":   5,
	"Dragonspine": 5,
})

var title = NewWeightedSlice(map[string]int{
	"None":                 200,
	"1st Fatui Harbinger":  10,
	"2nd Fatui Harbinger":  10,
	"3rd Fatui Harbinger":  10,
	"4th Fatui Harbinger":  10,
	"5th Fatui Harbinger":  10,
	"6th Fatui Harbinger":  10,
	"7th Fatui Harbinger":  10,
	"8th Fatui Harbinger":  10,
	"9th Fatui Harbinger":  10,
	"10th Fatui Harbinger": 10,
	"11th Fatui Harbinger": 10,
	"Adeptus":              70,
	"Archon":               70,
	"Sovereign":            50,
	"Hexenzirkel":          50,
	"Descender":            30,
	"The First Who Came":   5,
	"The Second Who Came":  5,
	"The Primordial One":   5,
	"Emanator":             1,
	"Herrscher":            1,
	"Aeon":                 1,
})

var outsideTeyvatTitles = map[string]struct{}{
	"Descender":           {},
	"The First Who Came":  {},
	"The Second Who Came": {},
	"The Primordial One":  {},
	"Emanator":            {},
	"Herrscher":           {},
	"Aeon":                {},
}

var models = NewWeightedSlice(map[string]int{
	"Tall male":     50,
	"Tall female":   50,
	"Medium male":   100,
	"Medium female": 100,
	"Short male":    40,
	"Short female":  40,
})

var visualAdjectives = NewWeightedSlice(map[string]int{
	"Boring":       10,
	"Elegant":      10,
	"Ferocious":    10,
	"Graceful":     10,
	"Mysterious":   10,
	"Sickly":       10,
	"Intimidating": 10,
	"Muscular":     10,
	"Fit":          10,
	"Cute":         10,
	"Soft":         10,
	"Skinny":       5,
	"Furry":        5,
	"Bulky":        5,
	"Brawny":       5,
	"Barefoot":     5,
	"Gloomy":       5,
	"Gothic":       5,
	"Stinky":       5,
	"Zombi":        3,
	"Chubby":       3,
	"Vtuber":       2,
})

var scaling = NewWeightedSlice(map[string]int{
	"ATK":             500,
	"HP":              200,
	"DEF":             100,
	"EM":              100,
	"Energy Recharge": 20,
	"EM and ATK":      20,
	"HP and ATK":      20,
	"DEF and ATK":     20,
	"Healing Bonus":   20,
	"Shield Strength": 5,
})

var roles = NewWeightedSlice(map[string]int{
	"On-field DPS":            10,
	"Off-field DPS":           7,
	"NA DPS":                  5,
	"Buffer":                  5,
	"Healer":                  5,
	"Shielder":                4,
	"Plunge DPS":              3,
	"Physical DPS":            3,
	"Healer DPS":              3,
	"Healer and shielder":     2,
	"Healer and shielder DPS": 1,
})

var strengths = NewWeightedSlice(map[string]int{
	"has good AOE":                              10,
	"excels in single-target damage":            10,
	"has good elemental application":            10,
	"it's a great battery":                      10,
	"easy to build":                             10,
	"has very high damage":                      10,
	"has great vertical scaling":                10,
	"has amazing animations and visual effects": 10,
	"can snapshot buffs":                        8,
	"provides strong team utility":              5,
	"can shred resistances":                     5,
	"can group enemies":                         5,
	"offers crowd control":                      4,
	"can heal while dealing damage":             3,
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
	"has low base stats":                          8,
	"has long skill animations":                   5,
	"has bad weapon options":                      5,
	"consumes a lot of stamina to play optimally": 5,
	"needs resistance to interruption to be good": 5,
	"doesn't create particles":                    5,
	"has very limited range":                      5,
	"can't crit":                                  2,
})

type GeneratedCharacter struct {
	name      string
	rarity    string
	element   string
	region    string
	weapon    string
	model     string
	adjective string
	scaling   string
	role      string
	strength  string
	weakness  string
	title     string
}

func (c GeneratedCharacter) PrettyString() string {
	return fmt.Sprintf(`%s is a %s %s character from %s.
Weapon: %s.
Model: %s %s.
Kit: %s, scales with %s, %s but %s.
Title: %s.`,
		c.name, c.rarity, c.element, c.region, c.weapon, c.adjective, c.model, c.role, c.scaling, c.strength, c.weakness, c.title)
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
	result.adjective = visualAdjectives.Random(rng)

	if result.rarity == "5*" || result.rarity == "6*" || result.rarity == "7*" {
		result.title = title.Random(rng)
	} else {
		result.title = "None"
	}

	_, outsideTeyvat := outsideTeyvatTitles[result.title]
	if outsideTeyvat {
		result.region = "outside Teyvat"
	} else {
		result.region = region.Random(rng)
	}

	return result
}

func generateSeedFromName(name string) int64 {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	hash := sha256.New()
	hash.Write([]byte(strings.ToLower(name)))
	return int64(binary.LittleEndian.Uint64(hash.Sum(nil)))
}
