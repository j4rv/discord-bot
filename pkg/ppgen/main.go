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
	"Ͼ": 80,
	"⋳": 80,
	"⋴": 60,
	"O": 60,
	"c": 40,
	"(": 20,
	"<": 10,
	"«": 5,
	"Ƈ": 3,
	"⟨": 1,
}

var rightPPHeads = map[string]int{
	"D": 100,
	"Ͽ": 80,
	"⋻": 80,
	"϶": 60,
	"O": 40,
	")": 20,
	">": 10,
	"»": 5,
	"ƿ": 5,
	"⟩": 1,
}

var leftPPBalls = map[string]int{
	"8": 100,
	"3": 80,
	"B": 60,
	"ᙣ": 30,
	"ɷ": 30,
	"ß": 20,
	"ɜ": 20,
	"ɞ": 10,
	"}": 10,
	"]": 10,
	"Ʒ": 5,
}

var rightPPBalls = map[string]int{
	"8": 100,
	"ᙦ": 30,
	"ɷ": 30,
	"E": 20,
	"ɛ": 20,
	"}": 10,
	"]": 10,
	"∑": 5,
	"Ƹ": 5,
}

var ppBodies = map[string]int{
	"=":   100,
	"≈":   50,
	"≍":   20,
	"≎":   20,
	"-":   5,
	"\\~": 5,
	"∾":   2,
	"≋":   1,
	"≭":   1,
}

var bigDickAscii1 = `
⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠛⢉⢉⠉⠉⠻⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⣿⠟⠠⡰⣕⣗⣷⣧⣀⣅⠘⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣿⠃⣠⣳⣟⣿⣿⣷⣿⡿⣜⠄⣿⣿⣿⣿⣿
⣿⣿⣿⣿⡿⠁⠄⣳⢷⣿⣿⣿⣿⡿⣝⠖⠄⣿⣿⣿⣿⣿
⣿⣿⣿⣿⠃⠄⢢⡹⣿⢷⣯⢿⢷⡫⣗⠍⢰⣿⣿⣿⣿⣿
⣿⣿⣿⡏⢀⢄⠤⣁⠋⠿⣗⣟⡯⡏⢎⠁⢸⣿⣿⣿⣿⣿
⣿⣿⣿⠄⢔⢕⣯⣿⣿⡲⡤⡄⡤⠄⡀⢠⣿⣿⣿⣿⣿⣿
⣿⣿⠇⠠⡳⣯⣿⣿⣾⢵⣫⢎⢎⠆⢀⣿⣿⣿⣿⣿⣿⣿
⣿⣿⠄⢨⣫⣿⣿⡿⣿⣻⢎⡗⡕⡅⢸⣿⣿⣿⣿⣿⣿⣿
⣿⣿⠄⢜⢾⣾⣿⣿⣟⣗⢯⡪⡳⡀⢸⣿⣿⣿⣿⣿⣿⣿
⣿⣿⠄⢸⢽⣿⣷⣿⣻⡮⡧⡳⡱⡁⢸⣿⣿⣿⣿⣿⣿⣿
⣿⣿⡄⢨⣻⣽⣿⣟⣿⣞⣗⡽⡸⡐⢸⣿⣿⣿⣿⣿⣿⣿
⣿⣿⡇⢀⢗⣿⣿⣿⣿⡿⣞⡵⡣⣊⢸⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⡀⡣⣗⣿⣿⣿⣿⣯⡯⡺⣼⠎⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣧⠐⡵⣻⣟⣯⣿⣷⣟⣝⢞⡿⢹⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⡆⢘⡺⣽⢿⣻⣿⣗⡷⣹⢩⢃⢿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣷⠄⠪⣯⣟⣿⢯⣿⣻⣜⢎⢆⠜⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⡆⠄⢣⣻⣽⣿⣿⣟⣾⡮⡺⡸⠸⣿⣿⣿⣿
⣿⣿⡿⠛⠉⠁⠄⢕⡳⣽⡾⣿⢽⣯⡿⣮⢚⣅⠹⣿⣿⣿
⡿⠋⠄⠄⠄⠄⢀⠒⠝⣞⢿⡿⣿⣽⢿⡽⣧⣳⡅⠌⠻⣿
⠁⠄⠄⠄⠄⠄⠐⡐⠱⡱⣻⡻⣝⣮⣟⣿⣻⣟⣻⡺⣊⠌`

var sussyDick = `
·------------ S U S S Y · D I C K ------------·
⣿⣿⠟⢹⣶⣶⣝⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⡟⢰⡌⠿⢿⣿⡾⢹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⢸⣿⣤⣒⣶⣾⣳⡻⣿⣿⣿⣿⡿⢛⣯⣭⣭⣭⣽⣻⣿⣿
⣿⣿⢸⣿⣿⣿⣿⢿⡇⣶⡽⣿⠟⣡⣶⣾⣯⣭⣽⣟⡻⣿⣷⡽
⣿⣿⠸⣿⣿⣿⣿⢇⠃⣟⣷⠃⢸⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣇⢻⣿⣿⣯⣕⠧⢿⢿⣇⢯⣝⣒⣛⣯⣭⣛⣛⣣⣿⣿⣿
⣿⣿⣿⣌⢿⣿⣿⣿⣿⡘⣞⣿⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣦⠻⠿⣿⣿⣷⠈⢞⡇⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⣿⣿⣗⠄⢿⣿⣿⡆⡈⣽⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿
⣿⣿⣿⡿⣻⣽⣿⣆⠹⣿⡇⠁⣿⡼⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟
⠿⣛⣽⣾⣿⣿⠿⠋⠄⢻⣷⣾⣿⣧⠟⣡⣾⣿⣿⣿⣿⣿⣿⡇
⡟⢿⣿⡿⠋⠁⣀⡀⠄⠘⠊⣨⣽⠁⠰⣿⣿⣿⣿⣿⣿⣿⡍⠗
⣿⠄⠄⠄⠄⣼⣿⡗⢠⣶⣿⣿⡇⠄⠄⣿⣿⣿⣿⣿⣿⣿⣇⢠
⣝⠄⠄⢀⠄⢻⡟⠄⣿⣿⣿⣿⠃⠄⠄⢹⣿⣿⣿⣿⣿⣿⣿⢹
⣿⣿⣿⣿⣧⣄⣁⡀⠙⢿⡿⠋⠄⣸⡆⠄⠻⣿⡿⠟⢛⣩⣝⣚
⣿⣿⣿⣿⣿⣿⣿⣿⣦⣤⣤⣤⣾⣿⣿⣄⠄⠄⠄⣴⣿⣿⣿⣇
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣄⡀⠛⠿⣿⣫⣾`

var sussyDick2 = `
⠀⠀⠀⠀⠀⢀⣴⡾⠿⠿⠿⠿⢶⣦⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⢠⣿⠁⠀⠀⠀⣀⣀⣀⣈⣻⣷⡄⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⣾⡇⠀⠀⣾⣟⠛⠋⠉⠉⠙⠛⢷⣄⠀⠀⠀⠀⠀⠀⠀
⢀⣤⣴⣶⣿⠀⠀⢸⣿⣿⣧⠀⠀⠀⠀⢀⣀⢹⡆⠀⠀⠀⠀⠀⠀
⢸⡏⠀⢸⣿⠀⠀⠀⢿⣿⣿⣷⣶⣶⣿⣿⣿⣿⠃⠀⠀⠀⠀⠀⠀
⣼⡇⠀⢸⣿⠀⠀⠀⠈⠻⠿⣿⣿⠿⠿⠛⢻⡇⠀⠀⠀⠀⠀⠀⠀
⣿⡇⠀⢸⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣤⣼⣷⣶⣶⣶⣤⡀⠀⠀
⣿⡇⠀⢸⣿⠀⠀⠀⠀⠀⠀⣀⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⡀
⢻⡇⠀⢸⣿⠀⠀⠀⠀⢀⣾⣿⣿⣿⣿⣿⣿⣿⡿⠿⣿⣿⣿⣿⡇
⠈⠻⠷⠾⣿⠀⠀⠀⠀⣾⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⢸⣿⣿⣿⣇
⠀⠀⠀⠀⣿⠀⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⠃⠀⢸⣿⣿⣿⡿
⠀⠀⠀⠀⢿⣧⣀⣠⣴⡿⠙⠛⠿⠿⠿⠿⠉⠀⠀⢠⣿⣿⣿⣿⠇
⠀⠀⠀⠀⠀⢈⣩⣭⣥⣤⣤⣤⣤⣤⣤⣤⣤⣤⣶⣿⣿⣿⣿⠏⠀
⠀⠀⠀⠀⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀
⠀⠀⠀⢸⣿⣿⣿⡟⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠋⠁⠀⠀⠀⠀
⠀⠀⠀⢸⣿⣿⣿⣷⣄⣀⣀⣀⣀⣀⣀⣀⣀⣀⡀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣦⡀⠀⠀⠀
⠀⠀⠀⠀⠀⠈⠛⠿⠿⣿⣿⣿⣿⣿⠿⠿⢿⣿⣿⣿⣿⣿⡄⠀⠀
⠀⠀⠀⠀⠀⠀⢀⣀⣀⣀⡀⠀⠀⠀⠀⠀⠀⢀⣹⣿⣿⣿⡇⠀⠀
⠀⠀⠀⠀⠀⢰⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠁⠀⠀
⠀⠀⠀⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠛⠁⠀⠀⠀
⠀⠀⠀⠀⣿⣿⣿⣿⠁⠀⠀⠀⠀⠀⠉⠉⠁⢤⣤⣤⣤⣤⣤⣤⡀
⠀⠀⠀⠀⢿⣿⣿⣿⣷⣶⣶⣶⣶⣾⣿⣿⣿⣆⢻⣿⣿⣿⣿⣿⡇
⠀⠀⠀⠀⠈⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⠻⣿⣿⣿⡿⠁
⠀⠀⠀⠀⠀⠀⠈⠙⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠉⠀⠙⠛⠉⠀⠀`

var ogreDick = `
·------------ O G R E · D I C K ------------·
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣀⠀⠀⣖⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠊⠀⠀⠔⡓⠊⡊⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡲⠤⠔⣡⡤⠄⠈⣁⣀⠱⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠁⠀⠐⠒⠊⢀⡤⠀⠱⢦⡤⣀⢄⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢤⢼⡀⠐⠶⠬⠵⠂⠀⡀⢀⡇⠀⡸⠙⠢⢄⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⡠⠺⠸⢻⡈⠢⢄⡀⠀⣀⠔⣠⠞⠺⢣⢡⠀⠀⠀⠑⣄⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⡜⠀⢠⢰⣇⠙⠦⡀⠠⢀⣶⡓⠁⠀⠈⠀⢸⠀⠀⠀⠀⠀⠢⠀
⠀⠀⠀⠀⠀⠀⠀⡜⠀⠀⠈⡇⢀⠀⡈⡟⠛⠃⢃⠤⠰⠦⠀⠠⠤⠲⣤⣜⠄⠀⠀⡇
⠀⠀⠀⠀⠀⠀⢰⠀⠀⠡⢰⣧⠽⠒⠒⠁⠀⠀⠀⠉⠉⠉⠉⠉⠉⢰⣻⠀⠀⠀⠀⡇
⠀⠀⠀⠀⠀⠀⠘⡀⠀⠈⠉⠂⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣇⠀⠀⠀⠰⠀
⠀⠀⠀⠀⠀⠀⠀⠑⠄⠀⠀⣆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣰⠉⠉⠐⠒⡇⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠈⠢⡎⢸⢦⣀⠀⠀⠀⠀⠀⠀⠀⣀⣠⣴⠾⠋⣆⣀⣀⡀⠇⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠛⠀⠙⠻⠶⠶⣶⡶⠶⠿⠛⠉⠀⠀⠀⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢳⢄⡀⠀⢀⣀⣠⣤⣀⠀⣀⡠⡔⠃⢸
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣤⣶⣾⣿⣿⣿⣿⣿⣧⠀⠀⠀⠀⠀⢸
⠀⠀⠀⠀⠀⠀⠀⣀⣤⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⢆⠀⠀⠀⠸
⠀⠀⠀⠀⢀⣴⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⢸⣀⣤⣼⠆
⢀⣴⣶⣝⢷⡝⢿⣿⣿⠿⠛⡤⠒⣰⣿⣿⢣⣿⣿⣿⣿⡇⢠⠋⠋⠁⠣⡀
⣼⣿⣿⣿⣿⣧⠻⡌⠋⠀⠀⠉⢰⣿⣿⡏⣸⣿⣿⣿⣿⣿⠘⠤⠀⠤⠔⠃
⠙⣿⣿⣿⡇⠋⠀⠀⠀⠀⠀⠀⠀⠈⠻⢿⠇⢻⣿⣿⣿⣿⡟
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠉⠉⠁
`

var paintDick = `
⠀⠖⠖⡆⠀⠀⠀⠀⣀⣀⣀⠀⠀
⢸⠀⠀⡗⠐⠉⠁⠀⠀⣇⡤⠽⡆
⠀⢉⡟⠳⡄⠀⠀⠀⢀⣇⣀⡴⠃
⠀⡏⠀⠀⡸⠉⠉⠉⠁⠀⠀⠀⠀
⠀⠙⠒⠚⠁⠀⠀⠀⠀⠀⠀⠀⠀
`

func NewPenis() string {
	return NewPenisWithSeed(time.Now().Unix())
}

func NewPenisWithSeed(seed int64) string {
	rng := rand.New(rand.NewSource(seed))
	superRareRng := rng.Float64()

	superRareTotalChance := 0.003
	superRareIndividualChance := superRareTotalChance / 5
	if superRareRng <= 1*superRareIndividualChance {
		return bigDickAscii1
	} else if superRareRng <= 2*superRareIndividualChance {
		return sussyDick
	} else if superRareRng <= 3*superRareIndividualChance {
		return sussyDick2
	} else if superRareRng <= 4*superRareIndividualChance {
		return ogreDick
	} else if superRareRng <= 5*superRareIndividualChance {
		return paintDick
	}

	if rng.Float64() <= 0.5 {
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
