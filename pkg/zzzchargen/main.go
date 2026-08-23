package zzzchargen

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"

	"github.com/j4rv/discord-bot/pkg/rngx"
)

var attributes = rngx.NewWeightedSlice(map[string]int{
	"Physical":   100,
	"Honed Edge": 10,
	"Fire":       100,
	"Ice":        100,
	"Frost":      10,
	"Electric":   100,
	"Ether":      100,
	"Auric Ink":  10,
	"Wind":       100,
	"Lumiflux":   10,

	"Solar": 2,
	"Lunar": 2,
	"Void":  1,
})

var specialties = rngx.NewWeightedSlice(map[string]int{
	"Attack":  100,
	"Stun":    100,
	"Anomaly": 100,
	"Support": 100,
	"Defense": 100,
	"Rupture": 100,
	"Armorer": 100,

	"Bangboomancer": 5,
	"Hybrid":        5,
})

var ranks = rngx.NewWeightedSlice(map[string]int{
	"S-Rank": 200,
	"A-Rank": 20,
	"∞-Rank": 5,
	"Z-Rank": 3,
	"C-Rank": 2,
	"B-Rank": 1,
})

var factions = rngx.NewWeightedSlice(map[string]int{
	"Angels of Delusion":                           100,
	"Belobog Heavy Industries":                     100,
	"Cunning Hares":                                100,
	"Criminal Investigation Special Response Team": 100,
	"Defense Force":                                100,
	"Hollow Special Operations Section 6":          100,
	"Hollow Raiders":                               100,
	"Mockingbird":                                  100,
	"New Eridu Public Security":                    100,
	"Obol Squad":                                   100,
	"Sons of Calydon":                              100,
	"Spook Shack":                                  100,
	"Stars of Lyra":                                100,
	"Victoria Housekeeping Co.":                    100,
	"Yunkui Summit":                                100,

	"White Star Institute": 5,
	"H.A.N.D":              5,
	"Independent":          5,
})

var combatStyles = rngx.NewWeightedSlice(map[string]int{
	"Sword":                     100,
	"Dual Swords":               80,
	"Dual Blades":               100,
	"Greatsword":                80,
	"Katana":                    60,
	"Dual Katanas":              30,
	"Rapier":                    40,
	"Saber":                     40,
	"Scimitar":                  20,
	"Combat Knife":              60,
	"Dual Knives":               40,
	"Throwing Knives":           20,
	"Cleaver":                   30,
	"Scythe":                    60,
	"Chainsaw":                  40,
	"Chain Blade":               20,
	"Blade Whip":                20,
	"Spear":                     80,
	"Halberd":                   30,
	"Glaive":                    30,
	"Staff":                     60,
	"Bo Staff":                  30,
	"Trident":                   20,
	"Poleaxe":                   20,
	"Gauntlets":                 100,
	"Brass Knuckles":            60,
	"Arm Blades":                40,
	"Claws":                     50,
	"Mechanical Fists":          30,
	"Hammer":                    50,
	"Warhammer":                 30,
	"Mace":                      30,
	"Club":                      20,
	"Baseball Bat":              20,
	"Metal Pipe":                10,
	"Shield":                    30,
	"Shield and Sword":          40,
	"Gun":                       100,
	"Dual Pistols":              100,
	"Revolver":                  40,
	"SMG":                       50,
	"Dual SMGs":                 30,
	"Shotgun":                   70,
	"Rifle":                     70,
	"Sniper Rifle":              30,
	"Machine Gun":               30,
	"Minigun":                   10,
	"Railgun":                   20,
	"Hand Cannon":               30,
	"Energy Cannon":             40,
	"Ether Cannon":              30,
	"Grenade Launcher":          20,
	"Flamethrower":              20,
	"Bow":                       40,
	"Compound Bow":              20,
	"Crossbow":                  20,
	"Bow and Blade":             30,
	"Thrown Weapons":            30,
	"Explosives":                20,
	"Grenades":                  20,
	"Cards":                     10,
	"Martial Arts":              60,
	"Kickboxing":                30,
	"Muay Thai":                 20,
	"Kung Fu":                   30,
	"Capoeira":                  15,
	"Boxing":                    30,
	"Wrestling":                 15,
	"Acrobatic Combat":          30,
	"Umbrella":                  40,
	"Parasol":                   30,
	"Fan":                       30,
	"War Fan":                   20,
	"Yo-yo":                     20,
	"Whip":                      30,
	"Chain Whip":                20,
	"Briefcase":                 20,
	"Suitcase Cannon":           20,
	"Wok":                       10,
	"Cooking Utensils":          10,
	"Musical Instrument":        20,
	"Combat Drones":             30,
	"Turrets":                   20,
	"Robotic Companion":         30,
	"Bangboo Mech Suit":         20,
	"Mech Suit":                 30,
	"Power Armor":               30,
	"Mechanical Arms":           20,
	"Cybernetic Limbs":          20,
	"Gun and Sword":             40,
	"Gun and Blade":             30,
	"Gun and Shield":            20,
	"Blade and Shield":          30,
	"Bow and Dual Blades":       20,
	"Gauntlets and Gun":         20,
	"Martial Arts and Gun":      20,
	"Motorcycle":                20,
	"Motorcycle and Blade":      15,
	"Talismans":                 20,
	"Paper Charms":              20,
	"Barrier Magic":             15,
	"Spatial Manipulation":      10,
	"Telekinesis":               10,
	"Fighting with a Bangboo":   5,
	"Weaponized Furniture":      3,
	"Weaponized Umbrella Stand": 1,
	"Pure Aura":                 2,
})

var models = rngx.NewWeightedSlice(map[string]int{
	"Tall male":     50,
	"Tall female":   50,
	"Medium male":   60,
	"Medium female": 60,
	"Short male":    50,
	"Short female":  50,

	"Beast Thiren": 20,
	"Machine":      20,
	"Other":        5,
})

var visualAdjectives = rngx.NewWeightedSlice(map[string]int{
	"Elegant":      10,
	"Ferocious":    10,
	"Stylish":      10,
	"Mysterious":   10,
	"Intimidating": 10,
	"Muscular":     10,
	"Cute":         10,
	"Cool":         10,
	"Chaotic":      10,
	"Cyberpunk":    10,
	"Boring":       10,
	"Graceful":     10,
	"Sickly":       10,
	"Fit":          10,
	"Soft":         10,
	"Furry":        10,
	"Gothic":       5,
	"Robotic":      5,
	"Angelic":      5,
	"Demonic":      5,
	"Sleepy":       5,
	"Skinny":       5,
	"Chubby":       5,
	"Bulky":        5,
	"Brawny":       5,
	"Barefoot":     5,
	"Gloomy":       5,
	"Edgy":         5,
	"Smug":         5,
	"Stinky":       5,
	"Dirty":        5,
	"Vtuber":       5,
	"Gremlin":      5,
	"Vampire":      5,
	"Zombi":        5,
	"Unhinged":     3,
	"Skeletal":     3,
	"Ghostly":      3,
})

var scaling = rngx.NewWeightedSlice(map[string]int{
	"ATK":                 150,
	"HP":                  100,
	"DEF":                 50,
	"Crit Rate":           100,
	"Crit Damage":         50,
	"PEN%":                50,
	"Impact":              100,
	"Anomaly Mastery":     100,
	"Anomaly Proficiency": 50,
	"Energy Regen":        100,
})

var strengths = rngx.NewWeightedSlice(map[string]int{
	"has good AOE":                        10,
	"excels in single-target damage":      10,
	"builds Daze extremely quickly":       10,
	"has good Attribute Anomaly buildup":  10,
	"buffs Disorder damage":               10,
	"has very high personal damage":       10,
	"has powerful Aftershocks":            10,
	"has Ether Veils":                     10,
	"has strong team buffs":               8,
	"has strong grouping":                 5,
	"has great mobility":                  5,
	"has a ton of invulnerability frames": 5,
	"can parry almost anything":           4,
	"has massive chain attack damage":     3,
	"has multiple ultimates":              3,
})

var weaknesses = rngx.NewWeightedSlice(map[string]int{
	"has energy issues":                   8,
	"is very difficult to play optimally": 8,
	"needs Mindscapes to feel complete":   8,
	"requires a lot of field time":        8,
	"is extremely fragile":                8,
	"has long animations":                 8,
	"has poor W-Engine options":           5,
	"has strict team requirements":        5,
	"struggles against mobile enemies":    5,
	"has shit AOE":                        5,
	"requires precise dodge timing":       5,
	"has awful Attribute buildup":         5,
})

var titles = rngx.NewWeightedSlice(map[string]int{
	"None": 250,

	"Void Hunter":                30,
	"The Hound":                  20,
	"The Tiger":                  20,
	"The Swordmaster":            20,
	"Champion of New Eridu":      20,
	"New Eridu's Finest":         30,
	"Chief":                      30,
	"Elite Investigator":         30,
	"Hollow Specialist":          30,
	"Master of the Hollow":       20,
	"Legendary Mercenary":        20,
	"Proxy":                      30,
	"The Uncrowned Lord":         15,
	"Thiren Champion":            15,
	"Champion of the Outer Ring": 15,
	"Professional Bangboo":       10,

	"The Night Watcher": 10,
	"The Last Sentinel": 10,
	"The Hollow Reaper": 10,
	"The Ether Phantom": 10,
	"The Stormbringer":  10,
	"The Unbroken":      10,
	"The Untamed":       10,
	"The Silent Blade":  10,
	"The Iron Fist":     10,
	"The Crimson Fang":  10,

	"Enchainer":             2,
	"Creator of New Eridu":  2,
	"Hollow Sovereign":      2,
	"The First Proxy":       2,
	"Administrator":         2,
	"The Voidwalker":        2,
	"Agent Zero":            2,
	"The End of the Hollow": 1,
})

var ratings = rngx.NewWeightedSlice(map[string]int{
	"-1/10": 2,
	"0/10":  10,
	"1/10":  20,
	"2/10":  25,
	"3/10":  30,
	"4/10":  40,
	"5/10":  50,
	"6/10":  60,
	"7/10":  70,
	"8/10":  60,
	"9/10":  40,
	"10/10": 30,
	"11/10": 2,
})

type GeneratedCharacter struct {
	name string

	rank      string
	attribute string
	specialty string

	faction     string
	combatStyle string
	model       string
	adjective   string

	scaling  string
	strength string
	weakness string

	title  string
	rating string

	simulatedDPS float64
}

func (c GeneratedCharacter) PrettyString() string {
	kit := fmt.Sprintf(
		"%s but %s.",
		c.strength,
		c.weakness,
	)

	if c.scaling != "" {
		kit = fmt.Sprintf(
			"scales with %s, %s but %s.",
			c.scaling,
			c.strength,
			c.weakness,
		)
	}

	return fmt.Sprintf(
		`%s is a %s %s %s agent affiliated with %s.
Combat Style: %s.
Model: %s %s.
Kit: %s
Title: %s.
Leaker Rating: %s.`,
		c.name,
		c.rank,
		c.attribute,
		c.specialty,
		c.faction,
		c.combatStyle,
		c.adjective,
		c.model,
		kit,
		c.title,
		c.rating,
	)
}

func NewChar(name string, seedSalt int64) GeneratedCharacter {
	var result GeneratedCharacter

	result.name = name

	rng := rand.New(
		rand.NewSource(generateSeedFromName(name) + seedSalt),
	)

	result.attribute = attributes.Random(rng)
	result.rank = ranks.Random(rng)
	result.specialty = specialties.Random(rng)

	result.faction = factions.Random(rng)
	result.combatStyle = combatStyles.Random(rng)

	result.model = models.Random(rng)
	result.adjective = visualAdjectives.Random(rng)

	switch result.specialty {
	case "Attack", "Anomaly", "Armorer", "Rupture":
		result.scaling = ""
	default:
		result.scaling = scaling.Random(rng)
	}

	result.strength = strengths.Random(rng)
	result.weakness = weaknesses.Random(rng)

	result.rating = ratings.Random(rng)

	if result.rank == "S-Rank" || result.rank == "Z-Rank" || result.rank == "∞-Rank" {
		result.title = titles.Random(rng)
	} else {
		result.title = "None"
	}

	return result
}

func generateSeedFromName(name string) int64 {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)

	hash := sha256.New()
	hash.Write([]byte(name))

	return int64(binary.LittleEndian.Uint64(hash.Sum(nil)))
}
