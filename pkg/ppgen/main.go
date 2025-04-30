package ppgen

import (
	"math/rand"
	"time"

	"github.com/j4rv/discord-bot/pkg/rngx"
)

const minLength = 1
const maxLength = 14

var leftPPHeads = rngx.NewWeightedSlice(map[string]int{
	"C": 100,
	"Ͼ": 60,
	"⋳": 60,
	"⋴": 60,
	"ϵ": 60,
	"O": 60,
	"c": 40,
	"८": 40,
	"(": 20,
	"<": 10,
	"«": 5,
	"Ƈ": 3,
	"⟨": 1,
})

var rightPPHeads = rngx.NewWeightedSlice(map[string]int{
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
})

var leftPPBalls = rngx.NewWeightedSlice(map[string]int{
	"8": 100,
	"3": 80,
	"Ɜ": 80,
	"B": 60,
	"฿": 50,
	"௰": 50,
	"ᙣ": 30,
	"ɷ": 30,
	"ß": 20,
	"ɜ": 20,
	"ɞ": 10,
	"෴": 10,
	"}": 10,
	"]": 10,
	"Ʒ": 5,
	"⧖": 1,
})

var rightPPBalls = rngx.NewWeightedSlice(map[string]int{
	"8": 100,
	"Ɛ": 80,
	"ᙦ": 30,
	"ɷ": 30,
	"E": 20,
	"ɛ": 20,
	"෴": 10,
	"}": 10,
	"]": 10,
	"∑": 5,
	"Ƹ": 5,
	"⧖": 1,
})

var ppBodies = rngx.NewWeightedSlice(map[string]int{
	"=":   100,
	"≈":   50,
	"≍":   20,
	"≎":   20,
	"≋":   5,
	"≔":   5,
	"≕":   5,
	"-":   5,
	"\\~": 5,
	"≭":   3,
	"≠":   3,
	"⋯":   2,
	"∾":   2,
	"∻":   1,
})

// 8╼╼╼╼╼D
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
	head := leftPPHeads.Random(rng)
	body := ppBodies.Random(rng)
	balls := leftPPBalls.Random(rng)

	penis := head
	for i := 0; i < length; i++ {
		penis += body
	}
	penis += balls

	return penis
}

func newPenisFacingRight(rng *rand.Rand) string {
	length := rng.Intn(maxLength-minLength) + minLength
	balls := rightPPBalls.Random(rng)
	body := ppBodies.Random(rng)
	head := rightPPHeads.Random(rng)

	penis := balls
	for i := 0; i < length; i++ {
		penis += body
	}
	penis += head

	return penis
}
