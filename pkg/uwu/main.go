package uwu

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"unicode"
)

var discordProtectedRegex = regexp.MustCompile(`<@!?\d+>|<a?:[a-zA-Z0-9_]+:\d+>|https?://[^\s]+`)

type uwuifier struct {
	stutterChance float64
	emoteChance   float64
	stretchChance float64
	suffixChance  float64
	emotes        []string
	suffixes      []string
	honorifics    []string
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
			"(⁄⁄>////<⁄⁄)",
			"(・`ω´・)",
			"(｡•̀⩊•́｡)",
			"(｡♥‿♥｡)",
			"(✿◠‿◠)",
			"(◕‿◕✿)",
			"(ᵕ•̤ᴗ•̤ᵕ)",
			"(╥﹏╥)",
			"(；ω ；)",
			"(ﾉ◕ヮ◕)ﾉ*:･ﾟ✧",
			"(づ￣ ³￣)づ",
			"(っ °Д °;)っ",
			"ヾ(*´∀ ˋ*)ﾉ",
			"(≧ω≦)",
			"(UwU)",
			"(>w<)",
			"(˘▾˘)",
			"( >д<)",
			"( • ̀ω•́ )",
			"(˶ᵔ ᵕ ᵔ˶)",
			"(´｡• ᵕ •｡`)",
			"(๑˃ᴗ˂)ﻭ",
			"(⁄˃ᆺ˂)⁄",
			"(：3 」∠ )",
			"(ᓀ‸ᓂ)",
			"(owo)",
			"owo",
			"OwO",
			"uwu",
			"UwU",
			"x3",
			":3",
			":3c",
			"\\*blushes\\*",
			"\\*faints\\*",
			"\\*hisses\\*",
			"\\*nuzzles\\*",
			"\\*paws you\\*",
			"\\*wiggles butt\\*",
			"\\*waggling my tail\\*",
			"\\*pounces on you\\*",
			"\\*climbs on your lap\\*",
			"\\*bites you\\*",
			"\\*smells you\\*",
			"\\*licking myself\\*",
			"\\*collapses\\*",
		},
		suffixes: []string{
			"~",
			"~~",
			"nya~",
			" rawr~",
			" meow~",
			" boop~",
			" murr~",
			" ehehe~",
			" purrr~",
		},
		honorifics: []string{
			"-chan",
			"-kun",
			"-san",
			"-sama",
			"-senpai",
			"-sensei",
			"-nyan",
		},
	}
}

func (u *uwuifier) UwUify(input string) string {
	if input == "" {
		return input
	}

	protected := make([]string, 0)
	input = discordProtectedRegex.ReplaceAllStringFunc(input, func(value string) string {
		if strings.HasPrefix(value, "<@") && len(u.honorifics) > 0 {
			value += u.honorifics[rand.Intn(len(u.honorifics))]
		}

		protected = append(protected, value)
		return fmt.Sprintf("__$%%&$%%&_%d__", len(protected)-1)
	})

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

	for i, value := range protected {
		output = strings.Replace(output, fmt.Sprintf("__$%%&$%%&_%d__", i), value, 1)
	}

	return output
}

func (u *uwuifier) replaceLetters(input string) string {
	var builder strings.Builder

	for _, char := range input {
		switch char {
		case 'r':
			builder.WriteRune('w')
		case 'R':
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
	parts := regexp.MustCompile(`\S+|\s+`).FindAllString(input, -1)

	var candidates []int
	for i, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}

		runes := []rune(part)
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
	runes := []rune(parts[index])
	last := runes[len(runes)-1]
	parts[index] += strings.Repeat(string(last), rand.Intn(3)+1)
	return strings.Join(parts, "")
}

func (u *uwuifier) addRandomSuffixes(input string) string {
	parts := regexp.MustCompile(`\S+|\s+`).FindAllString(input, -1)

	for i, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}

		if len([]rune(part)) <= 4 || rand.Float64() >= u.suffixChance || len(u.suffixes) == 0 {
			continue
		}

		suffix := u.suffixes[rand.Intn(len(u.suffixes))]

		end := len(part)
		for end > 0 {
			last := part[end-1]
			if last != '!' && last != '?' && last != '.' && last != ',' {
				break
			}
			end--
		}

		parts[i] = part[:end] + suffix + part[end:]
	}

	return strings.Join(parts, "")
}
