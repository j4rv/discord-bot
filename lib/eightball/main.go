package eightball

import "math/rand"

var responses = []string{
	// "yes"
	"Yes.",
	"It is certain.",
	"It is decidedly so.",
	"Without a doubt.",
	"Yes definitely.",
	"You may rely on it.",
	"As I see it, yes.",
	"Most likely.",
	"Outlook good.",
	"Signs point to yes.",

	// "neutral"
	"Reply hazy, try again.",
	"Ask again later.",
	"Better not tell you now.",
	"Cannot predict now.",
	"Concentrate and ask again.",
	"Stop asking questions.",
	"Trust me, you don't want to hear the answer.",
	"Why do you want to know?",
	"It's fifty fifty.",
	"I don't get paid enough to answer that.",

	// "no"
	"No.",
	"Don't count on it.",
	"My reply is no.",
	"My sources say no.",
	"Outlook not so good.",
	"Very doubtful.",
	"It is as likely as winning the lottery.",
	"I don't think so.",
	"The opposite of yes.",
	"Nay, nope, nein, non.",
}

func Response() string {
	index := rand.Intn(len(responses))
	return responses[index]
}
