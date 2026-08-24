package uwu

import (
	"math/rand"
	"strings"
	"unicode"
)

type uwuifier struct {
	stutterChance float64
	emoteChance   float64
	stretchChance float64
	suffixChance  float64
	emotes        []string
	suffixes      []string
}

func NewUwuifier() *uwuifier {
	return &uwuifier{
		stutterChance: 0.35,
		emoteChance:   0.70,
		stretchChance: 0.25,
		suffixChance:  0.08,
		emotes: []string{
			"(≧◡≦)",
			"(⁄ ⁄•⁄ω⁄•⁄ ⁄)",
			"(・`ω´・)",
			"(｡♥‿♥｡)",
			"(✿◠‿◠)",
			"(◕‿◕✿)",
			"(ᵕ•̤ᴗ•̤ᵕ)",
			"(ﾉ◕ヮ◕)ﾉ*:･ﾟ✧",
			"(≧ω≦)",
			"(UwU)",
			"(>w<)",
			"(˶ᵔ ᵕ ᵔ˶)",
			"(´｡• ᵕ •｡`)",
			"(๑˃ᴗ˂)ﻭ",
			"(owo)",
			"owo",
			"OwO",
			"uwu",
			"UwU",
			":3",
			":3c",
			"(blushes)",
			"(faints)",
			"(wiggling my tail)",
		},
		suffixes: []string{
			"~",
			"~~",
			"nya~",
			" rawr~",
		},
	}
}

func (u *uwuifier) UwUify(input string) string {
	if input == "" {
		return input
	}

	output := u.replaceLetters(input)
	output = u.addRandomSuffixes(output)

	if rand.Float64() < u.stutterChance {
		output = u.addStutter(output)
	}

	if rand.Float64() < u.stretchChance {
		output = u.stretchRandomWord(output)
	}

	if rand.Float64() < u.emoteChance && len(u.emotes) > 0 {
		output += " " + u.emotes[rand.Intn(len(u.emotes))]
	}

	return output
}

func (u *uwuifier) replaceLetters(input string) string {
	var builder strings.Builder

	for _, char := range input {
		switch char {
		case 'r', 'l':
			builder.WriteRune('w')
		case 'R', 'L':
			builder.WriteRune('W')
		case '!', '?':
			builder.WriteRune(char)
			builder.WriteRune(char)
		default:
			builder.WriteRune(char)
		}
	}

	return builder.String()
}

func (u *uwuifier) addStutter(input string) string {
	for i, char := range input {
		if !unicode.IsLetter(char) {
			continue
		}

		count := rand.Intn(2) + 1
		stutter := strings.Repeat(string(char)+"-", count)

		return input[:i] + stutter + input[i:]
	}

	return input
}

func (u *uwuifier) stretchRandomWord(input string) string {
	words := strings.Fields(input)

	var candidates []int
	for i, word := range words {
		runes := []rune(word)
		if len(runes) < 2 {
			continue
		}

		last := runes[len(runes)-1]
		if strings.ContainsRune("aeiouAEIOU", last) {
			candidates = append(candidates, i)
		}
	}

	if len(candidates) == 0 {
		return input
	}

	index := candidates[rand.Intn(len(candidates))]
	runes := []rune(words[index])
	last := runes[len(runes)-1]

	words[index] += strings.Repeat(string(last), rand.Intn(3)+1)

	return strings.Join(words, " ")
}

func (u *uwuifier) addRandomSuffixes(input string) string {
	words := strings.Fields(input)

	for i, word := range words {
		// dont add suffix to very short words
		if len([]rune(word)) <= 4 || rand.Float64() >= u.suffixChance || len(u.suffixes) == 0 {
			continue
		}

		suffix := u.suffixes[rand.Intn(len(u.suffixes))]

		// Put the suffix before trailing punctuation.
		end := len(word)
		for end > 0 {
			last := word[end-1]
			if last != '!' && last != '?' && last != '.' && last != ',' {
				break
			}
			end--
		}

		words[i] = word[:end] + suffix + word[end:]
	}

	return strings.Join(words, " ")
}
