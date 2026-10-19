package eightball

import (
	"math/rand"
	"strings"
	"unicode"
)

var singleWordResponseTemplates = []string{
	`You really asked "<word>" to Jarv Bot. Incredible.
I'm a fortune teller, not your English teacher.
https://dictionary.cambridge.org/dictionary/english/<word>`,

	`ERROR 404: Actual question not found.
Received input: "<word>"
https://dictionary.cambridge.org/dictionary/english/<word>`,

	`Bro really typed "<word>" and expected divine intervention.
Here: https://dictionary.cambridge.org/dictionary/english/<word>
Even the spirits can't help you with this one.`,

	`I'm Jarv Bot, not a walking dictionary.
Sigh, fine: https://dictionary.cambridge.org/dictionary/english/<word>
Please allocate at least one more brain cell to this question.`,

	`The 8-Ball has reviewed "<word>" and is disappointed.
https://dictionary.cambridge.org/dictionary/english/<word>
Please try again after unlocking the ability to form sentences.`,

	`New achievement unlocked: Asking absolutely nothing!
Reward: One free trip to the dictionary.
https://dictionary.cambridge.org/dictionary/english/<word>`,

	`"<word>"? Be serious for a second.
I'm trying to predict the future, not teach vocabulary.
https://dictionary.cambridge.org/dictionary/english/<word>`,

	`Question length: 1 word.
Effort detected: 0.
Critical thinking: not found.
https://dictionary.cambridge.org/dictionary/english/<word>`,

	`The spirits have reviewed your input.
They unanimously voted to send you back to school.
https://dictionary.cambridge.org/dictionary/english/<word>`,

	`You used "<word>"! It's not very effective...
Unfortunately, neither is your ability to ask questions.
https://dictionary.cambridge.org/dictionary/english/<word>`,

	`My Magic 8-Ball is confused.
And honestly, "<word>" isn't giving it much to work with.
https://dictionary.cambridge.org/dictionary/english/<word>`,

	`"<word>" ... fascinating.
Unfortunately, my crystal ball doesn't come with a dictionary.
https://dictionary.cambridge.org/dictionary/english/<word>`,

	`Jarv Bot has left the server.
Reason: Forced to process "<word>".
https://dictionary.cambridge.org/dictionary/english/<word>`,

	`Congratulations! You've successfully typed one word.
Your reward is absolutely no answer from me.
Just kidding, here's your actual reward: https://dictionary.cambridge.org/dictionary/english/<word>`,

	`I can predict the future, not read your mind.
What am I supposed to do with "<word>", exactly?
This is all I got: https://dictionary.cambridge.org/dictionary/english/<word>`,

	`The answer is unclear.
The question is also unclear. Actually, where's the question?
https://dictionary.cambridge.org/dictionary/english/<word>`,
}

func isSingleWordQuestion(question string) bool {
	words := strings.Fields(question)
	return len(words) == 1
}

func singleWordQuestionResponse(word string) string {
	cleanedWord := strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return -1
		}
		return r
	}, word)

	if cleanedWord == "" {
		return "You managed to send absolutely nothing. Honestly, impressive. That shouldn't even be possible :ohno:."
	}

	template := singleWordResponseTemplates[rand.Intn(len(singleWordResponseTemplates))]
	template = strings.ReplaceAll(template, "<word>", cleanedWord)
	return template
}
