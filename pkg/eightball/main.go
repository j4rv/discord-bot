package eightball

import "math/rand"

var yesResponses = []string{
	"Yes.",
	"Yes! 🎉",
	"Sí.",
	"It is certain.",
	"It is decidedly so.",
	"Without a doubt.",
	"Yes definitely.",
	"You may rely on it.",
	"As I see it, yes.",
	"Most likely.",
	"Outlook good.",
	"Signs point to yes.",
	"The stars align.",
	"Absolutely.",
	"I'm afraid so.",
	"*nods*",
	"Yes, yes, yes, yes!",
	"Yes... eventually.",
}

var noResponses = []string{
	"No.",
	"No ❤️",
	"Nay, nope, nein, non.",
	"Nah.",
	"Don't count on it.",
	"My reply is no.",
	"My sources say no.",
	"Outlook not so good.",
	"Very doubtful.",
	"It is as likely as winning the lottery.",
	"I don't think so.",
	"The opposite of yes.",
	"In your dreams.",
	"Not in this universe.",
	"Thankfully no.",
	"How can I say this... No.",
	"No, no, no, no!",
	"Calculating chance... updating weather patterns... checking the Shadow Realm...\nOutcome: 0% chance.",
}

var neutralResponses = []string{
	"Reply hazy, try again.",
	"Ask again later.",
	"Better not tell you now.",
	"Cannot predict now.",
	"Concentrate and ask again.",
	"Stop asking questions.",
	"Trust me, you don't want to hear the answer.",
	"Why do you want to know?",
	"It is fifty fifty.",
	"That depends.",
	"I don't get paid enough to answer that.",
	"Fuck you.",
	"Cu... come again?",
	"Please stop pinging me",
	"Why are you like this?",
	"Ask someone else.",
	"I'll pretend you didn't ask that question.",
	"That question will keep me up at night.",
	"For legal reasons, I refuse to answer that question.",
	"404 Answer Not Found.",
	"¯\\_(ツ)_/¯",
}

func Response() string {
	rng := rand.Float64()
	neutralChance := 0.10
	yesChance := 0.45

	if rng < neutralChance {
		index := rand.Intn(len(neutralResponses))
		return neutralResponses[index]
	}

	if rng < neutralChance+yesChance {
		index := rand.Intn(len(yesResponses))
		return yesResponses[index]
	}

	index := rand.Intn(len(noResponses))
	return noResponses[index]
}
